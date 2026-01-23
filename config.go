package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"dnsres/instrumentation"
)

// Duration wraps time.Duration to support human-friendly strings in JSON
// (e.g., "5s", "1m").
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		dur, err := time.ParseDuration(s)
		if err != nil {
			return err
		}
		d.Duration = dur
		return nil
	}
	// fallback: try as number (nanoseconds)
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		d.Duration = time.Duration(n)
		return nil
	}
	return json.Unmarshal(b, &d.Duration)
}

// Config represents the configuration for the DNS resolver
type Config struct {
	Hostnames            []string `json:"hostnames"`
	DNSServers           []string `json:"dns_servers"`
	QueryTimeout         Duration `json:"query_timeout"`
	QueryInterval        Duration `json:"query_interval"`
	HealthPort           int      `json:"health_port"`
	MetricsPort          int      `json:"metrics_port"`
	LogDir               string   `json:"log_dir"`
	InstrumentationLevel string   `json:"instrumentation_level"`
	CircuitBreaker       struct {
		Threshold int      `json:"threshold"`
		Timeout   Duration `json:"timeout"`
	} `json:"circuit_breaker"`
	Cache struct {
		MaxSize int64 `json:"max_size"`
	} `json:"cache"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Hostnames) == 0 {
		return fmt.Errorf("no hostnames specified")
	}
	if len(c.DNSServers) == 0 {
		return fmt.Errorf("no DNS servers specified")
	}
	if c.QueryTimeout.Duration <= 0 {
		return fmt.Errorf("invalid query timeout")
	}
	if c.QueryInterval.Duration <= 0 {
		return fmt.Errorf("invalid query interval")
	}
	if c.CircuitBreaker.Threshold <= 0 {
		return fmt.Errorf("invalid circuit breaker threshold")
	}
	if c.CircuitBreaker.Timeout.Duration <= 0 {
		return fmt.Errorf("invalid circuit breaker timeout")
	}
	if c.Cache.MaxSize <= 0 {
		return fmt.Errorf("invalid cache max size")
	}
	if _, err := instrumentation.ParseLevel(c.InstrumentationLevel); err != nil {
		return fmt.Errorf("invalid instrumentation level: %w", err)
	}
	return nil
}

// loadConfig loads the configuration from a file
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}
	config.InstrumentationLevel = normalizeInstrumentationLevel(config.InstrumentationLevel)

	// Ensure DNS servers have ports
	for i, server := range config.DNSServers {
		if _, _, err := net.SplitHostPort(server); err != nil {
			config.DNSServers[i] = net.JoinHostPort(server, "53")
		}
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	return &config, nil
}

// validateConfig validates the configuration values
func validateConfig(cfg *Config) error {
	if len(cfg.Hostnames) == 0 {
		return errors.New("at least one hostname must be specified")
	}
	if len(cfg.DNSServers) == 0 {
		return errors.New("at least one DNS server must be specified")
	}
	if cfg.QueryTimeout.Duration <= 0 {
		return errors.New("query timeout must be positive")
	}
	if cfg.QueryInterval.Duration <= 0 {
		return errors.New("query interval must be positive")
	}
	if cfg.CircuitBreaker.Threshold <= 0 {
		return errors.New("circuit breaker threshold must be positive")
	}
	if cfg.CircuitBreaker.Timeout.Duration <= 0 {
		return errors.New("circuit breaker timeout must be positive")
	}
	if cfg.Cache.MaxSize <= 0 {
		return errors.New("cache max size must be positive")
	}
	if _, err := instrumentation.ParseLevel(cfg.InstrumentationLevel); err != nil {
		return fmt.Errorf("invalid instrumentation level: %w", err)
	}
	return nil
}

func normalizeInstrumentationLevel(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}
	return strings.ToLower(strings.TrimSpace(value))
}
