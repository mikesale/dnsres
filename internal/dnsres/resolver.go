package dnsres

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"dnsres/cache"
	"dnsres/circuitbreaker"
	"dnsres/dnsanalysis"
	"dnsres/dnspool"
	"dnsres/health"
	"dnsres/instrumentation"
	"dnsres/metrics"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// DNSResolver represents a DNS resolution tool
type DNSResolver struct {
	config                *Config
	clientPool            *dnspool.ClientPool
	breakers              map[string]*circuitbreaker.CircuitBreaker
	cache                 *cache.ShardedCache
	health                *health.HealthChecker
	successLog            *log.Logger
	errorLog              *log.Logger
	appLog                *log.Logger
	output                io.Writer
	stats                 *ResolutionStats
	instrumentationLevel  instrumentation.Level
	resolveAllFunc        func(context.Context)
	resolveWithServerFunc func(context.Context, string, string) (*dnsanalysis.DNSResponse, error)
	getClient             func(string) (dnsClient, error)
	putClient             func(string, dnsClient)
	events                *eventBus
	logDir                string
	logDirFallback        bool
}

type dnsClient interface {
	ExchangeContext(context.Context, *dns.Msg, string) (*dns.Msg, time.Duration, error)
}

// NewDNSResolver creates a new DNS resolver
func NewDNSResolver(config *Config) (*DNSResolver, error) {
	config.InstrumentationLevel = normalizeInstrumentationLevel(config.InstrumentationLevel)
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize loggers
	successLog, errorLog, appLog, actualLogDir, wasFallback, err := setupLoggers(config.LogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to setup loggers: %w", err)
	}

	// Initialize client pool
	clientPool := dnspool.NewClientPool(100, config.QueryTimeout.Duration)

	// Initialize circuit breakers
	breakers := make(map[string]*circuitbreaker.CircuitBreaker)
	for _, server := range config.DNSServers {
		breakers[server] = circuitbreaker.NewCircuitBreaker(
			config.CircuitBreaker.Threshold,
			config.CircuitBreaker.Timeout.Duration,
			server,
		)
	}

	// Initialize sharded cache
	cache := cache.NewShardedCache(config.Cache.MaxSize, 16)

	// Initialize health checker
	level, err := instrumentation.ParseLevel(config.InstrumentationLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid instrumentation level: %w", err)
	}

	healthChecker := health.NewHealthChecker(config.DNSServers, appLog, level)

	// Initialize stats
	stats := &ResolutionStats{
		StartTime: time.Now(),
		Stats:     make(map[string]*ServerStats),
	}
	for _, server := range config.DNSServers {
		stats.Stats[server] = &ServerStats{}
	}

	resolver := &DNSResolver{
		config:                config,
		clientPool:            clientPool,
		breakers:              breakers,
		cache:                 cache,
		health:                healthChecker,
		successLog:            successLog,
		errorLog:              errorLog,
		appLog:                appLog,
		output:                os.Stdout,
		stats:                 stats,
		instrumentationLevel:  level,
		resolveAllFunc:        nil,
		resolveWithServerFunc: nil,
		getClient:             nil,
		putClient:             nil,
		events:                newEventBus(),
		logDir:                actualLogDir,
		logDirFallback:        wasFallback,
	}
	resolver.resolveAllFunc = resolver.resolveAll
	resolver.resolveWithServerFunc = resolver.resolveWithServer
	resolver.getClient = func(server string) (dnsClient, error) {
		return clientPool.Get(server)
	}
	resolver.putClient = func(server string, client dnsClient) {
		dnsClient, ok := client.(*dns.Client)
		if !ok {
			return
		}
		clientPool.Put(server, dnsClient)
	}

	resolver.appLogf(
		instrumentation.Low,
		"resolver initialized hostnames=%d servers=%d interval=%s timeout=%s instrumentation=%s",
		len(config.Hostnames),
		len(config.DNSServers),
		config.QueryInterval.Duration,
		config.QueryTimeout.Duration,
		level.String(),
	)

	return resolver, nil
}

// SubscribeEvents returns a channel of resolver activity events.
func (r *DNSResolver) SubscribeEvents(buffer int) (<-chan ResolverEvent, func()) {
	if r.events == nil {
		return nil, func() {}
	}
	return r.events.subscribe(buffer)
}

// SetOutputWriter controls where resolver status output is written.
func (r *DNSResolver) SetOutputWriter(writer io.Writer) {
	r.output = writer
}

// HealthSnapshot returns the latest health check status.
func (r *DNSResolver) HealthSnapshot() map[string]bool {
	if r.health == nil {
		return map[string]bool{}
	}
	return r.health.StatusSnapshot()
}

// GetLogDir returns the actual log directory being used.
func (r *DNSResolver) GetLogDir() string {
	return r.logDir
}

// LogDirWasFallback returns true if the log directory fell back to $HOME/logs.
func (r *DNSResolver) LogDirWasFallback() bool {
	return r.logDirFallback
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

	r.outputf("Health endpoint listening on :%d\n", r.config.HealthPort)
	r.outputf("Metrics endpoint listening on :%d\n", r.config.MetricsPort)
	r.appLogf(instrumentation.Low, "health server starting on :%d", r.config.HealthPort)
	r.appLogf(instrumentation.Low, "metrics server starting on :%d", r.config.MetricsPort)

	// Start servers
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.appLog.Printf("Health server error: %v", err)
		}
	}()
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.appLog.Printf("Metrics server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := healthServer.Shutdown(shutdownCtx); err != nil {
			r.appLog.Printf("Health server shutdown error: %v", err)
		}
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			r.appLog.Printf("Metrics server shutdown error: %v", err)
		}
	}()

	// Start resolution loop
	r.resolveAllFunc(ctx) // Run initial resolution immediately
	r.outputf("Resolution loop started (interval %s)\n", r.config.QueryInterval.Duration)
	r.appLogf(instrumentation.Low, "resolution loop started interval=%s", r.config.QueryInterval.Duration)

	ticker := time.NewTicker(r.config.QueryInterval.Duration)
	defer ticker.Stop()

	return r.runLoop(ctx, ticker.C)
}

func (r *DNSResolver) runLoop(ctx context.Context, ticks <-chan time.Time) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticks:
			r.appLogf(instrumentation.Low, "resolution tick fired interval=%s", r.config.QueryInterval.Duration)
			r.resolveAllFunc(ctx)
		}
	}
}

// resolveAll resolves all hostnames against all DNS servers concurrently
func (r *DNSResolver) resolveAll(ctx context.Context) {
	start := time.Now()
	r.outputf("Resolution cycle starting (hostnames %d, servers %d)\n", len(r.config.Hostnames), len(r.config.DNSServers))
	r.emitEvent(ResolverEvent{
		Type:          EventCycleStart,
		Time:          start,
		HostnameCount: len(r.config.Hostnames),
		ServerCount:   len(r.config.DNSServers),
	})
	r.appLogf(
		instrumentation.Low,
		"resolution cycle start hostnames=%d servers=%d",
		len(r.config.Hostnames),
		len(r.config.DNSServers),
	)

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
					response, err := r.resolveWithServerFunc(ctx, s, h)
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
					consistentValue := false
					r.emitEvent(ResolverEvent{
						Type:       EventInconsistent,
						Time:       time.Now(),
						Hostname:   h,
						Consistent: &consistentValue,
					})
					r.appLogf(instrumentation.High, "inconsistent responses hostname=%s", h)
					r.errorLog.Printf("Inconsistent responses for %s", h)
				}
			}
		}(hostname)
	}
	wg.Wait()
	duration := time.Since(start)
	metrics.DNSResolutionCycleDuration.Observe(duration.Seconds())
	r.outputf("Resolution cycle complete (duration %s)\n", duration)
	r.emitEvent(ResolverEvent{
		Type:          EventCycleComplete,
		Time:          time.Now(),
		Duration:      duration,
		HostnameCount: len(r.config.Hostnames),
		ServerCount:   len(r.config.DNSServers),
	})
	r.appLogf(instrumentation.Low, "resolution cycle complete duration=%s", duration)
}

// resolveWithServer resolves a hostname using a specific DNS server
func (r *DNSResolver) resolveWithServer(ctx context.Context, server, hostname string) (*dnsanalysis.DNSResponse, error) {
	// Check cache first
	if cached, ok := r.cache.Get(hostname); ok {
		metrics.DNSResolutionCacheHit.WithLabelValues(server, hostname).Inc()
		r.appLogf(instrumentation.Low, "cache hit hostname=%s server=%s", hostname, server)
		r.emitEvent(ResolverEvent{
			Type:      EventResolveSuccess,
			Time:      time.Now(),
			Hostname:  hostname,
			Server:    server,
			Addresses: append([]string(nil), cached.Addresses...),
			Source:    "cache",
		})
		return cached, nil
	}
	metrics.DNSResolutionCacheMiss.WithLabelValues(server, hostname).Inc()
	r.appLogf(instrumentation.Low, "cache miss hostname=%s server=%s", hostname, server)

	// Check circuit breaker
	if !r.breakers[server].Allow() {
		metrics.DNSResolutionFailure.WithLabelValues(server, hostname, "circuit_breaker").Inc()
		r.appLogf(instrumentation.Medium, "circuit breaker open server=%s", server)
		r.emitEvent(ResolverEvent{
			Type:     EventResolveFailure,
			Time:     time.Now(),
			Hostname: hostname,
			Server:   server,
			Error:    "circuit breaker open",
			Source:   "circuit_breaker",
		})
		return nil, fmt.Errorf("circuit breaker open for %s", server)
	}

	// Get client from pool
	client, err := r.getClient(server)
	if err != nil {
		r.appLogf(instrumentation.Medium, "client pool get failed server=%s err=%v", server, err)
		r.emitEvent(ResolverEvent{
			Type:     EventResolveFailure,
			Time:     time.Now(),
			Hostname: hostname,
			Server:   server,
			Error:    err.Error(),
			Source:   "client_pool",
		})
		return nil, fmt.Errorf("failed to get client from pool: %w", err)
	}
	defer r.putClient(server, client)

	// Create DNS message
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(hostname), dns.TypeA)
	msg.RecursionDesired = true
	msg.SetEdns0(4096, true) // Enable EDNS with DNSSEC

	// Increment total resolution attempts
	metrics.DNSResolutionTotal.WithLabelValues(server, hostname).Inc()

	// Send query
	start := time.Now()
	response, _, err := client.ExchangeContext(ctx, msg, server)
	elapsed := time.Since(start)

	if err != nil {
		r.breakers[server].RecordFailure()
		r.stats.Stats[server].Failures++
		r.stats.Stats[server].LastError = err.Error()
		metrics.DNSResolutionFailure.WithLabelValues(server, hostname, "query_error").Inc()
		r.appLogf(instrumentation.Medium, "DNS query failed hostname=%s server=%s err=%v", hostname, server, err)
		r.emitEvent(ResolverEvent{
			Type:     EventResolveFailure,
			Time:     time.Now(),
			Hostname: hostname,
			Server:   server,
			Duration: elapsed,
			Error:    err.Error(),
			Source:   "query_error",
		})
		return nil, fmt.Errorf("DNS query failed: %w", err)
	}

	// Record metrics
	duration := elapsed.Seconds()
	metrics.DNSResolutionDuration.WithLabelValues(server, hostname).Observe(duration)
	metrics.DNSResponseSize.WithLabelValues(server, hostname).Observe(float64(response.Len()))

	// Process response
	if response.Rcode != dns.RcodeSuccess {
		r.breakers[server].RecordFailure()
		r.stats.Stats[server].Failures++
		r.stats.Stats[server].LastError = dns.RcodeToString[response.Rcode]
		metrics.DNSResolutionFailure.WithLabelValues(server, hostname, dns.RcodeToString[response.Rcode]).Inc()
		r.appLogf(
			instrumentation.Medium,
			"DNS response error hostname=%s server=%s rcode=%s",
			hostname,
			server,
			dns.RcodeToString[response.Rcode],
		)
		r.emitEvent(ResolverEvent{
			Type:     EventResolveFailure,
			Time:     time.Now(),
			Hostname: hostname,
			Server:   server,
			Duration: elapsed,
			Error:    dns.RcodeToString[response.Rcode],
			Source:   "rcode",
		})
		return nil, fmt.Errorf("DNS query returned error code: %s", dns.RcodeToString[response.Rcode])
	}

	r.breakers[server].RecordSuccess()
	r.stats.Stats[server].Total++
	metrics.DNSResolutionSuccess.WithLabelValues(server, hostname).Inc()
	r.appLogf(
		instrumentation.High,
		"DNS response ok hostname=%s server=%s duration=%s",
		hostname,
		server,
		elapsed,
	)

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

	r.emitEvent(ResolverEvent{
		Type:      EventResolveSuccess,
		Time:      time.Now(),
		Hostname:  hostname,
		Server:    server,
		Duration:  elapsed,
		Addresses: append([]string(nil), dnsResponse.Addresses...),
		Source:    "query",
	})

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

// Helper functions
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func (r *DNSResolver) appLogf(level instrumentation.Level, format string, args ...any) {
	if r.appLog == nil || r.instrumentationLevel < level {
		return
	}
	r.appLog.Printf(format, args...)
}

func (r *DNSResolver) emitEvent(event ResolverEvent) {
	if r.events == nil {
		return
	}
	r.events.publish(event)
}

func (r *DNSResolver) outputf(format string, args ...any) {
	if r.output == nil {
		return
	}
	fmt.Fprintf(r.output, format, args...)
}
