package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
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

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	reportMode := flag.Bool("report", false, "Generate statistics report")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	// Load configuration
	fmt.Printf("Loading configuration from %s\n", *configFile)
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println("Configuration loaded")

	// Override hostname if specified
	if *hostname != "" {
		config.Hostnames = []string{*hostname}
		fmt.Printf("Hostname override enabled: %s\n", *hostname)
	}

	// Create resolver
	fmt.Println("Validating configuration")
	resolver, err := NewDNSResolver(config)
	if err != nil {
		log.Fatalf("Failed to create DNS resolver: %v", err)
	}
	fmt.Println("Resolver initialized")

	// Handle report mode
	if *reportMode {
		fmt.Println("Report mode enabled; generating report")
		fmt.Println(resolver.GenerateReport())
		return
	}

	fmt.Printf("Monitoring %d hostnames across %d DNS servers every %s\n", len(config.Hostnames), len(config.DNSServers), config.QueryInterval.Duration)
	fmt.Println("Press q then Enter to quit")

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("Shutdown signal received (%s)\n", sig)
		cancel()
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.TrimSpace(scanner.Text())
			if strings.EqualFold(input, "q") {
				fmt.Println("Quit requested; shutting down")
				cancel()
				return
			}
		}
	}()

	// Start resolution
	if err := resolver.Start(ctx); err != nil {
		log.Fatalf("Failed to start DNS resolver: %v", err)
	}
}
