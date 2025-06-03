package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"dnsres/cache"
	"dnsres/circuitbreaker"
	"dnsres/dnsanalysis"
	"dnsres/dnspool"
	"dnsres/health"
	"dnsres/metrics"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config represents the configuration for the DNS resolver
type Config struct {
	Hostnames        []string      `json:"hostnames"`
	DNSServers       []string      `json:"dns_servers"`
	QueryTimeout     time.Duration `json:"query_timeout"`
	QueryInterval    time.Duration `json:"query_interval"`
	FailureThreshold int           `json:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
	HalfOpenTimeout  time.Duration `json:"half_open_timeout"`
	MaxCacheTTL      time.Duration `json:"max_cache_ttl"`
	HealthPort       int           `json:"health_port"`
	MetricsPort      int           `json:"metrics_port"`
	LogDir           string        `json:"log_dir"`
	CircuitBreaker   struct {
		Threshold int           `json:"threshold"`
		Timeout   time.Duration `json:"timeout"`
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
	if c.QueryTimeout <= 0 {
		return fmt.Errorf("invalid query timeout")
	}
	if c.CircuitBreaker.Threshold <= 0 {
		return fmt.Errorf("invalid circuit breaker threshold")
	}
	if c.CircuitBreaker.Timeout <= 0 {
		return fmt.Errorf("invalid circuit breaker timeout")
	}
	if c.Cache.MaxSize <= 0 {
		return fmt.Errorf("invalid cache max size")
	}
	return nil
}

// ResolutionStats tracks resolution statistics
type ResolutionStats struct {
	Total     int
	Failures  int
	LastError string
	StartTime time.Time
	Stats     map[string]*ServerStats
}

// ServerStats tracks statistics for a single server
type ServerStats struct {
	Total     int
	Failures  int
	LastError string
}

// setupLoggers initializes the loggers
func setupLoggers(logDir string) (*log.Logger, *log.Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	successLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-success.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open success log file: %w", err)
	}

	errorLogFile, err := os.OpenFile(
		filepath.Join(logDir, "dnsres-error.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open error log file: %w", err)
	}

	successLog := log.New(successLogFile, "", log.LstdFlags)
	errorLog := log.New(errorLogFile, "", log.LstdFlags)

	return successLog, errorLog, nil
}

// DNSResolver represents a DNS resolution tool
type DNSResolver struct {
	config     *Config
	clientPool *dnspool.ClientPool
	breakers   map[string]*circuitbreaker.CircuitBreaker
	cache      *cache.ShardedCache
	health     *health.HealthChecker
	successLog *log.Logger
	errorLog   *log.Logger
	stats      *ResolutionStats
}

// NewDNSResolver creates a new DNS resolver
func NewDNSResolver(config *Config) (*DNSResolver, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize loggers
	successLog, errorLog, err := setupLoggers(config.LogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to setup loggers: %w", err)
	}

	// Initialize client pool
	clientPool := dnspool.NewClientPool(100, config.QueryTimeout)

	// Initialize circuit breakers
	breakers := make(map[string]*circuitbreaker.CircuitBreaker)
	for _, server := range config.DNSServers {
		breakers[server] = circuitbreaker.NewCircuitBreaker(
			config.CircuitBreaker.Threshold,
			config.CircuitBreaker.Timeout,
		)
	}

	// Initialize sharded cache
	cache := cache.NewShardedCache(config.Cache.MaxSize, 16)

	// Initialize health checker
	healthChecker := health.NewHealthChecker(config.DNSServers)

	// Initialize stats
	stats := &ResolutionStats{
		StartTime: time.Now(),
		Stats:     make(map[string]*ServerStats),
	}
	for _, server := range config.DNSServers {
		stats.Stats[server] = &ServerStats{}
	}

	return &DNSResolver{
		config:     config,
		clientPool: clientPool,
		breakers:   breakers,
		cache:      cache,
		health:     healthChecker,
		successLog: successLog,
		errorLog:   errorLog,
		stats:      stats,
	}, nil
}

// Start begins the DNS resolution monitoring
func (r *DNSResolver) Start(ctx context.Context) error {
	// Create HTTP servers
	healthServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", r.config.HealthPort),
		Handler:      r.health,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	metricsServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", r.config.MetricsPort),
		Handler:      promhttp.Handler(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start servers
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.errorLog.Printf("Health server error: %v", err)
		}
	}()
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.errorLog.Printf("Metrics server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			r.errorLog.Printf("Health server shutdown error: %v", err)
		}
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			r.errorLog.Printf("Metrics server shutdown error: %v", err)
		}
	}()

	// Start resolution loop
	ticker := time.NewTicker(r.config.QueryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			r.resolveAll(ctx)
		}
	}
}

// resolveAll resolves all hostnames against all DNS servers concurrently
func (r *DNSResolver) resolveAll(ctx context.Context) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent resolutions

	for _, hostname := range r.config.Hostnames {
		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			var responses []*dnsanalysis.DNSResponse
			var responseMu sync.Mutex

			// Resolve against all servers concurrently
			var serverWg sync.WaitGroup
			for _, server := range r.config.DNSServers {
				serverWg.Add(1)
				go func(s string) {
					defer serverWg.Done()
					response, err := r.resolveWithServer(ctx, s, h)
					if err != nil {
						r.errorLog.Printf("Failed to resolve %s using %s: %v", h, s, err)
						r.stats.Stats[s].Failures++
						r.stats.Stats[s].LastError = err.Error()
						return
					}
					r.successLog.Printf("Resolved %s using %s (state: %s)", h, s, r.breakers[s].GetState())
					r.stats.Stats[s].Total++

					responseMu.Lock()
					responses = append(responses, response)
					responseMu.Unlock()
				}(server)
			}
			serverWg.Wait()

			// Check response consistency
			if len(responses) > 1 {
				consistent := dnsanalysis.CompareResponses(responses)
				metrics.DNSResolutionConsistency.WithLabelValues(h).Set(boolToFloat64(consistent))
				if !consistent {
					r.errorLog.Printf("Inconsistent responses for %s", h)
				}
			}
		}(hostname)
	}
	wg.Wait()
}

// resolveWithServer resolves a hostname using a specific DNS server
func (r *DNSResolver) resolveWithServer(ctx context.Context, server, hostname string) (*dnsanalysis.DNSResponse, error) {
	// Check cache first
	if cached, ok := r.cache.Get(hostname); ok {
		metrics.DNSCacheHits.Inc()
		return cached, nil
	}
	metrics.DNSCacheMisses.Inc()

	// Check circuit breaker
	if !r.breakers[server].Allow() {
		metrics.DNSCircuitBreakerTrips.WithLabelValues(server).Inc()
		return nil, fmt.Errorf("circuit breaker open for %s", server)
	}

	// Get client from pool
	client, err := r.clientPool.Get(server)
	if err != nil {
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer r.clientPool.Put(client)

	// Create DNS message
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeA)

	// Send query
	start := time.Now()
	response, _, err := client.ExchangeWithContext(ctx, msg, server)
	if err != nil {
		r.breakers[server].RecordFailure()
		r.stats.Stats[server].Failures++
		r.stats.Stats[server].LastError = err.Error()
		return nil, fmt.Errorf("DNS query failed: %w", err)
	}

	// Record metrics
	duration := time.Since(start).Seconds()
	metrics.DNSResolutionDuration.WithLabelValues(server).Observe(duration)
	metrics.DNSResponseSize.WithLabelValues(server).Observe(float64(response.Len()))

	// Process response
	if response.Rcode != dns.RcodeSuccess {
		r.breakers[server].RecordFailure()
		r.stats.Stats[server].Failures++
		r.stats.Stats[server].LastError = dns.RcodeToString[response.Rcode]
		return nil, fmt.Errorf("DNS query returned error code: %s", dns.RcodeToString[response.Rcode])
	}

	r.breakers[server].RecordSuccess()
	r.stats.Stats[server].Total++

	// Create DNS response
	ttl := getMinTTL(response)
	dnsResponse := &dnsanalysis.DNSResponse{
		Server:    server,
		Hostname:  hostname,
		Addresses: make([]string, 0),
		TTL:       ttl,
	}

	// Extract IP addresses
	for _, answer := range response.Answer {
		if a, ok := answer.(*dns.A); ok {
			dnsResponse.Addresses = append(dnsResponse.Addresses, a.A.String())
		}
	}

	// Cache the response
	r.cache.Set(hostname, dnsResponse, time.Duration(ttl)*time.Second)

	return dnsResponse, nil
}

// getMinTTL returns the minimum TTL from a DNS response
func getMinTTL(msg *dns.Msg) uint32 {
	if len(msg.Answer) == 0 {
		return 0
	}

	minTTL := msg.Answer[0].Header().Ttl
	for _, rr := range msg.Answer {
		if rr.Header().Ttl < minTTL {
			minTTL = rr.Header().Ttl
		}
	}
	return minTTL
}

// GenerateReport generates a statistics report
func (r *DNSResolver) GenerateReport() string {
	var report strings.Builder
	report.WriteString("Hour              | DNS Server     | Total    | Fails    | Fail %  \n")
	report.WriteString("-----------------------------------------------------------------\n")

	// Sort stats by time
	keys := make([]string, 0, len(r.stats.Stats))
	for k := range r.stats.Stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		stats := r.stats.Stats[key]
		server := key
		hour := r.stats.StartTime.Format("2006-01-02 15:04")
		failPercent := float64(stats.Failures) / float64(stats.Total) * 100
		report.WriteString(fmt.Sprintf("%s | %-12s | %-8d | %-8d | %6.2f%%\n",
			hour, server, stats.Total, stats.Failures, failPercent))
	}

	return report.String()
}

// Helper functions
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	reportMode := flag.Bool("report", false, "Generate statistics report")
	hostname := flag.String("host", "", "Override hostname from config file")
	flag.Parse()

	// Load configuration
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override hostname if specified
	if *hostname != "" {
		config.Hostnames = []string{*hostname}
	}

	// Create resolver
	resolver, err := NewDNSResolver(config)
	if err != nil {
		log.Fatalf("Failed to create DNS resolver: %v", err)
	}

	// Handle report mode
	if *reportMode {
		fmt.Println(resolver.GenerateReport())
		return
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Start resolution
	if err := resolver.Start(ctx); err != nil {
		log.Fatalf("Failed to start DNS resolver: %v", err)
	}
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
	if cfg.QueryTimeout <= 0 {
		return errors.New("query timeout must be positive")
	}
	if cfg.QueryInterval <= 0 {
		return errors.New("query interval must be positive")
	}
	if cfg.FailureThreshold <= 0 {
		return errors.New("failure threshold must be positive")
	}
	if cfg.ResetTimeout <= 0 {
		return errors.New("reset timeout must be positive")
	}
	if cfg.HalfOpenTimeout <= 0 {
		return errors.New("half-open timeout must be positive")
	}
	if cfg.MaxCacheTTL <= 0 {
		return errors.New("max cache TTL must be positive")
	}
	if cfg.HealthPort <= 0 || cfg.HealthPort > 65535 {
		return errors.New("invalid health port")
	}
	if cfg.MetricsPort <= 0 || cfg.MetricsPort > 65535 {
		return errors.New("invalid metrics port")
	}
	if cfg.LogDir == "" {
		cfg.LogDir = "logs"
	}
	return nil
}
