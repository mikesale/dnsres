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
func TestResolveConfigPath(t *testing.T) {
	tests := []struct {
		name         string
		explicitPath string
		setupFunc    func(t *testing.T, tempDir string)
		wantContains string // substring that should be in the path
		wantCreated  bool
		wantEmpty    bool // expect empty string return
	}{
		{
			name:         "explicit path provided",
			explicitPath: "/custom/path/config.json",
			wantContains: "/custom/path/config.json",
			wantCreated:  false,
		},
		{
			name:         "local config.json exists",
			explicitPath: "",
			setupFunc: func(t *testing.T, tempDir string) {
				// Create ./config.json in current directory
				configJSON := []byte(`{
  "hostnames": ["test.com"],
  "dns_servers": ["8.8.8.8:53"],
  "query_timeout": "5s",
  "query_interval": "30s",
  "circuit_breaker": {"threshold": 5, "timeout": "30s"},
  "cache": {"max_size": 1000}
}`)
				oldWd, _ := os.Getwd()
				defer os.Chdir(oldWd)
				os.Chdir(tempDir)
				if err := os.WriteFile("config.json", configJSON, 0644); err != nil {
					t.Fatalf("failed to create ./config.json: %v", err)
				}
			},
			wantContains: "config.json",
			wantCreated:  false,
		},
		{
			name:         "XDG config auto-created",
			explicitPath: "",
			setupFunc: func(t *testing.T, tempDir string) {
				// Set XDG_CONFIG_HOME to temp dir
				os.Setenv("XDG_CONFIG_HOME", tempDir)
			},
			wantContains: "dnsres/config.json",
			wantCreated:  true,
		},
		{
			name:         "XDG config already exists",
			explicitPath: "",
			setupFunc: func(t *testing.T, tempDir string) {
				// Create XDG config first
				os.Setenv("XDG_CONFIG_HOME", tempDir)
				configDir := filepath.Join(tempDir, "dnsres")
				os.MkdirAll(configDir, 0755)
				configJSON := []byte(`{
  "hostnames": ["existing.com"],
  "dns_servers": ["1.1.1.1:53"],
  "query_timeout": "5s",
  "query_interval": "30s",
  "circuit_breaker": {"threshold": 5, "timeout": "30s"},
  "cache": {"max_size": 1000}
}`)
				os.WriteFile(filepath.Join(configDir, "config.json"), configJSON, 0644)
			},
			wantContains: "dnsres/config.json",
			wantCreated:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)

			tempDir := t.TempDir()

			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			gotPath, gotCreated, err := ResolveConfigPath(tt.explicitPath)

			if err != nil {
				t.Fatalf("ResolveConfigPath returned error: %v", err)
			}

			if tt.wantEmpty {
				if gotPath != "" {
					t.Errorf("expected empty path, got %s", gotPath)
				}
				return
			}

			if !strings.Contains(gotPath, tt.wantContains) {
				t.Errorf("expected path to contain %q, got %q", tt.wantContains, gotPath)
			}

			if gotCreated != tt.wantCreated {
				t.Errorf("expected wasCreated=%v, got %v", tt.wantCreated, gotCreated)
			}

			// If created, verify the file exists and is valid JSON
			if gotCreated && gotPath != "" {
				if _, err := os.Stat(gotPath); err != nil {
					t.Errorf("created config file doesn't exist: %v", err)
				}
				// Try to load it to verify it's valid
				if _, err := LoadConfig(gotPath); err != nil {
					t.Errorf("created config file is invalid: %v", err)
				}
			}
		})
	}
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("local config.json takes precedence over XDG", func(t *testing.T) {
		oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)

		tempDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tempDir)

		// Create both ./config.json and XDG config
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tempDir)

		localConfig := []byte(`{"hostnames": ["local.com"], "dns_servers": ["8.8.8.8:53"], "query_timeout": "5s", "query_interval": "30s", "circuit_breaker": {"threshold": 5, "timeout": "30s"}, "cache": {"max_size": 1000}}`)
		os.WriteFile("config.json", localConfig, 0644)

		xdgDir := filepath.Join(tempDir, "dnsres")
		os.MkdirAll(xdgDir, 0755)
		xdgConfig := []byte(`{"hostnames": ["xdg.com"], "dns_servers": ["1.1.1.1:53"], "query_timeout": "5s", "query_interval": "30s", "circuit_breaker": {"threshold": 5, "timeout": "30s"}, "cache": {"max_size": 1000}}`)
		os.WriteFile(filepath.Join(xdgDir, "config.json"), xdgConfig, 0644)

		gotPath, _, _ := ResolveConfigPath("")

		// Should use ./config.json
		if !strings.HasSuffix(gotPath, "config.json") || strings.Contains(gotPath, "dnsres") {
			t.Errorf("expected local config.json, got %s", gotPath)
		}

		// Verify it loads the local config
		cfg, _ := LoadConfig(gotPath)
		if cfg.Hostnames[0] != "local.com" {
			t.Errorf("expected local.com from local config, got %s", cfg.Hostnames[0])
		}
	})

	t.Run("explicit path overrides everything", func(t *testing.T) {
		tempDir := t.TempDir()
		explicitPath := filepath.Join(tempDir, "custom.json")

		customConfig := []byte(`{"hostnames": ["custom.com"], "dns_servers": ["9.9.9.9:53"], "query_timeout": "5s", "query_interval": "30s", "circuit_breaker": {"threshold": 5, "timeout": "30s"}, "cache": {"max_size": 1000}}`)
		os.WriteFile(explicitPath, customConfig, 0644)

		gotPath, _, _ := ResolveConfigPath(explicitPath)

		if gotPath != explicitPath {
			t.Errorf("expected %s, got %s", explicitPath, gotPath)
		}
	})
}

func TestConfigPathEdgeCases(t *testing.T) {
	t.Run("config file is a directory not a file", func(t *testing.T) {
		tempDir := t.TempDir()
		configDir := filepath.Join(tempDir, "config.json")
		os.Mkdir(configDir, 0755)

		// This should fail when trying to load, not in ResolveConfigPath
		_, err := LoadConfig(configDir)
		if err == nil {
			t.Error("expected error when loading directory as config")
		}
	})

	t.Run("config file has invalid JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		os.WriteFile(configPath, []byte("not valid json {{{"), 0644)

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("config file with no read permissions", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		os.WriteFile(configPath, []byte(`{"hostnames":["test.com"]}`), 0000)
		defer os.Chmod(configPath, 0644) // cleanup

		_, err := LoadConfig(configPath)
		if err == nil {
			t.Error("expected error for unreadable file")
		}
	})
}

func TestResolverGetLogDir(t *testing.T) {
	tests := []struct {
		name         string
		logDir       string
		setupFunc    func(t *testing.T) string // returns expected path substring
		wantContains string
	}{
		{
			name:   "custom log directory",
			logDir: "/custom/logs",
			setupFunc: func(t *testing.T) string {
				tempDir := t.TempDir()
				customDir := filepath.Join(tempDir, "custom", "logs")
				os.MkdirAll(customDir, 0755)
				return customDir
			},
			wantContains: "custom/logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logDirPath string
			if tt.setupFunc != nil {
				logDirPath = tt.setupFunc(t)
			} else {
				logDirPath = tt.logDir
			}

			config := &Config{
				Hostnames:     []string{"example.com"},
				DNSServers:    []string{"8.8.8.8:53"},
				QueryTimeout:  Duration{Duration: 5 * time.Second},
				QueryInterval: Duration{Duration: 30 * time.Second},
				LogDir:        logDirPath,
			}
			config.CircuitBreaker.Threshold = 1
			config.CircuitBreaker.Timeout = Duration{Duration: 30 * time.Second}
			config.Cache.MaxSize = 10

			resolver, err := NewDNSResolver(config)
			if err != nil {
				t.Fatalf("failed to create resolver: %v", err)
			}

			gotLogDir := resolver.GetLogDir()
			if !strings.Contains(gotLogDir, tt.wantContains) {
				t.Errorf("expected log dir to contain %q, got %q", tt.wantContains, gotLogDir)
			}
		})
	}
}

func TestResolverLogDirWasFallback(t *testing.T) {
	t.Run("returns false for custom log directory", func(t *testing.T) {
		tempDir := t.TempDir()
		customDir := filepath.Join(tempDir, "custom-logs")
		os.MkdirAll(customDir, 0755)

		config := &Config{
			Hostnames:     []string{"example.com"},
			DNSServers:    []string{"8.8.8.8:53"},
			QueryTimeout:  Duration{Duration: 5 * time.Second},
			QueryInterval: Duration{Duration: 30 * time.Second},
			LogDir:        customDir,
		}
		config.CircuitBreaker.Threshold = 1
		config.CircuitBreaker.Timeout = Duration{Duration: 30 * time.Second}
		config.Cache.MaxSize = 10

		resolver, err := NewDNSResolver(config)
		if err != nil {
			t.Fatalf("failed to create resolver: %v", err)
		}

		if resolver.LogDirWasFallback() {
			t.Error("expected LogDirWasFallback to return false for custom directory")
		}
	})
}
