package dnsres

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dnsres/cache"
	"dnsres/dnsanalysis"
)

func TestValidateConfigInstrumentationLevel(t *testing.T) {
	base := Config{
		Hostnames:     []string{"example.com"},
		DNSServers:    []string{"8.8.8.8:53"},
		QueryTimeout:  Duration{Duration: 5 * time.Second},
		QueryInterval: Duration{Duration: 30 * time.Second},
		HealthPort:    8880,
		MetricsPort:   9990,
		LogDir:        "logs",
	}
	base.CircuitBreaker.Threshold = 1
	base.CircuitBreaker.Timeout = Duration{Duration: 30 * time.Second}
	base.Cache.MaxSize = 10

	valid := base
	valid.InstrumentationLevel = "HiGh"
	if err := validateConfig(&valid); err != nil {
		t.Fatalf("expected valid instrumentation level, got error: %v", err)
	}

	invalid := base
	invalid.InstrumentationLevel = "verbose"
	if err := validateConfig(&invalid); err == nil {
		t.Fatalf("expected error for invalid instrumentation level")
	}
}

func TestLoadConfigNormalizesDNSServerPorts(t *testing.T) {
	configJSON := []byte(`{
  "hostnames": ["example.com"],
  "dns_servers": ["8.8.8.8", "1.1.1.1:54"],
  "query_timeout": "5s",
  "query_interval": "30s",
  "health_port": 8880,
  "metrics_port": 9990,
  "log_dir": "logs",
  "circuit_breaker": {
    "threshold": 1,
    "timeout": "30s"
  },
  "cache": {
    "max_size": 10
  }
}`)

	configPath := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if len(cfg.DNSServers) != 2 {
		t.Fatalf("expected 2 DNS servers, got %d", len(cfg.DNSServers))
	}
	if cfg.DNSServers[0] != "8.8.8.8:53" {
		t.Fatalf("expected normalized server, got %s", cfg.DNSServers[0])
	}
	if cfg.DNSServers[1] != "1.1.1.1:54" {
		t.Fatalf("expected existing port preserved, got %s", cfg.DNSServers[1])
	}
}

func TestResolveWithServerUsesCache(t *testing.T) {
	entry := &dnsanalysis.DNSResponse{Hostname: "example.com"}
	shardedCache := cache.NewShardedCache(1024, 1)
	shardedCache.Set("example.com", entry, time.Minute)

	resolver := &DNSResolver{cache: shardedCache}
	got, err := resolver.resolveWithServer(context.Background(), "8.8.8.8:53", "example.com")
	if err != nil {
		t.Fatalf("resolveWithServer returned error: %v", err)
	}
	if got != entry {
		t.Fatalf("expected cached response, got %+v", got)
	}
}

func TestNormalizeInstrumentationLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "   ",
			expected: "none",
		},
		{
			name:     "lowercase",
			input:    "LoW",
			expected: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeInstrumentationLevel(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	stats := &ResolutionStats{
		StartTime: time.Date(2025, 1, 2, 15, 4, 0, 0, time.UTC),
		Stats: map[string]*ServerStats{
			"8.8.8.8:53": {Total: 10, Failures: 2},
			"1.1.1.1:53": {Total: 0, Failures: 0},
		},
	}
	resolver := &DNSResolver{stats: stats}

	report := resolver.GenerateReport()
	if !strings.Contains(report, "8.8.8.8:53") {
		t.Fatalf("expected report to include server, got %s", report)
	}
	if !strings.Contains(report, "20.00%") {
		t.Fatalf("expected failure percentage in report, got %s", report)
	}
	if !strings.Contains(report, "1.1.1.1:53") {
		t.Fatalf("expected report to include zero-total server, got %s", report)
	}
	if !strings.Contains(report, "  0.00%") {
		t.Fatalf("expected zero percent for empty totals, got %s", report)
	}
}
