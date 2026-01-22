package main

import (
	"testing"
	"time"
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
