package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	// Basic Information
	Timestamp     time.Time `json:"timestamp"`
	Level         string    `json:"level"`
	Hostname      string    `json:"hostname"`
	Server        string    `json:"server"`
	CorrelationID string    `json:"correlation_id"`

	// System Context
	Version     string `json:"version"`
	Environment string `json:"environment"`
	InstanceID  string `json:"instance_id"`

	// DNS Query Details
	QueryType        string `json:"query_type"`
	EDNSEnabled      bool   `json:"edns_enabled"`
	DNSSECEnabled    bool   `json:"dnssec_enabled"`
	RecursionDesired bool   `json:"recursion_desired"`

	// Performance Metrics
	Duration       float64 `json:"duration_ms,omitempty"`
	QueueTime      float64 `json:"queue_time_ms,omitempty"`
	NetworkLatency float64 `json:"network_latency_ms,omitempty"`
	ProcessingTime float64 `json:"processing_time_ms,omitempty"`
	CacheTTL       int64   `json:"cache_ttl_seconds,omitempty"`

	// Response Analysis
	ResponseCode  string   `json:"response_code,omitempty"`
	ResponseSize  int      `json:"response_size,omitempty"`
	RecordCount   int      `json:"record_count,omitempty"`
	Authoritative bool     `json:"authoritative,omitempty"`
	Truncated     bool     `json:"truncated,omitempty"`
	ResponseFlags []string `json:"response_flags,omitempty"`

	// Circuit Breaker and Cache
	CircuitState string `json:"circuit_state"`
	CacheHit     bool   `json:"cache_hit,omitempty"`

	// Error Information
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"error_type,omitempty"`
}

// setupLoggers initializes the loggers
func setupLoggers(logDir string) (*log.Logger, *log.Logger, *log.Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	successLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-success.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open success log file: %w", err)
	}

	errorLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-error.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open error log file: %w", err)
	}

	appLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-app.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open app log file: %w", err)
	}

	successLog := log.New(successLogFile, "", log.LstdFlags)
	errorLog := log.New(errorLogFile, "", log.LstdFlags)
	appLog := log.New(appLogFile, "", log.LstdFlags)

	return successLog, errorLog, appLog, nil
}
