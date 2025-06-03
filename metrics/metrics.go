package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DNSResolutionTotal tracks total DNS resolution attempts
	DNSResolutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_total",
			Help: "Total number of DNS resolution attempts",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionSuccess tracks successful DNS resolutions
	DNSResolutionSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_success",
			Help: "Number of successful DNS resolutions",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionFailure tracks failed DNS resolutions
	DNSResolutionFailure = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_failure",
			Help: "Number of failed DNS resolutions",
		},
		[]string{"server", "hostname", "error_type"},
	)

	// DNSResolutionDuration tracks DNS resolution duration
	DNSResolutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_duration_seconds",
			Help:    "DNS resolution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"server", "hostname"},
	)

	// CircuitBreakerState tracks the current state of circuit breakers
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of each DNS server's circuit breaker (0=Closed, 1=Open, 2=Half-Open)",
		},
		[]string{"server"},
	)

	// CircuitBreakerFailures tracks consecutive failures
	CircuitBreakerFailures = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_failures",
			Help: "Number of consecutive failures for each DNS server",
		},
		[]string{"server"},
	)

	// DNSResponseSize tracks the size of DNS responses
	DNSResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_response_size_bytes",
			Help:    "Size of DNS responses in bytes",
			Buckets: []float64{64, 128, 256, 512, 1024, 2048, 4096},
		},
		[]string{"server", "hostname"},
	)

	// DNSRecordCount tracks the number of records in responses
	DNSRecordCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_record_count",
			Help:    "Number of records in DNS responses",
			Buckets: []float64{1, 2, 5, 10, 20, 50, 100},
		},
		[]string{"server", "hostname", "record_type"},
	)

	// DNSResolutionLatency tracks the latency between servers
	DNSResolutionLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_latency_seconds",
			Help:    "Latency between different DNS servers",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"server1", "server2", "hostname"},
	)

	// DNSResolutionConsistency tracks response consistency
	DNSResolutionConsistency = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_consistency",
			Help: "Whether responses from different DNS servers are consistent (1 = consistent, 0 = inconsistent)",
		},
		[]string{"hostname"},
	)

	// DNSResolutionTTL tracks TTL values from responses
	DNSResolutionTTL = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_ttl_seconds",
			Help:    "TTL values from DNS responses",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"server", "hostname", "record_type"},
	)

	// DNSResolutionRetries tracks retry attempts
	DNSResolutionRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_retries_total",
			Help: "Total number of DNS resolution retry attempts",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionTimeout tracks timeout occurrences
	DNSResolutionTimeout = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_timeout_total",
			Help: "Total number of DNS resolution timeouts",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionNXDOMAIN tracks NXDOMAIN responses
	DNSResolutionNXDOMAIN = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_nxdomain_total",
			Help: "Total number of NXDOMAIN responses",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionSERVFAIL tracks SERVFAIL responses
	DNSResolutionSERVFAIL = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_servfail_total",
			Help: "Total number of SERVFAIL responses",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionRefused tracks REFUSED responses
	DNSResolutionRefused = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_refused_total",
			Help: "Total number of REFUSED responses",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionRateLimit tracks rate limiting occurrences
	DNSResolutionRateLimit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_rate_limit_total",
			Help: "Total number of rate limit occurrences",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionNetworkError tracks network-related errors
	DNSResolutionNetworkError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_network_error_total",
			Help: "Total number of network-related errors",
		},
		[]string{"server", "hostname", "error_type"},
	)

	// DNSResolutionDNSSEC tracks DNSSEC validation results
	DNSResolutionDNSSEC = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_dnssec_total",
			Help: "Total number of DNSSEC validation results",
		},
		[]string{"server", "hostname", "status"},
	)

	// DNSResolutionEDNS tracks EDNS support
	DNSResolutionEDNS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_edns_support",
			Help: "EDNS support status (1=supported, 0=not supported)",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionDNSSECSupport tracks DNSSEC support
	DNSResolutionDNSSECSupport = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_dnssec_support",
			Help: "DNSSEC support status (1=supported, 0=not supported)",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionProtocol tracks protocol usage
	DNSResolutionProtocol = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_protocol_total",
			Help: "Total number of DNS resolutions by protocol",
		},
		[]string{"server", "hostname", "protocol"},
	)

	// DNSResolutionCacheHit tracks cache hits
	DNSResolutionCacheHit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_cache_hit_total",
			Help: "Total number of DNS cache hits",
		},
		[]string{"server", "hostname"},
	)

	// DNSResolutionCacheMiss tracks cache misses
	DNSResolutionCacheMiss = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_cache_miss_total",
			Help: "Total number of DNS cache misses",
		},
		[]string{"server", "hostname"},
	)

	// Cache metrics
	CacheSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "dns_resolver_cache_size",
			Help: "Current number of entries in the DNS cache",
		},
	)
	CacheHits = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dns_resolver_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)
	CacheMisses = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dns_resolver_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)
	CacheEvictions = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dns_resolver_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
	)

	// Health check metrics
	HealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolver_health_status",
			Help: "Health status of the DNS resolver (1 = healthy, 0 = unhealthy)",
		},
		[]string{"component"},
	)
	HealthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolver_health_check_duration_seconds",
			Help:    "Duration of health checks in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"component"},
	)
)

func init() {
	// Register DNS resolution metrics
	prometheus.MustRegister(DNSResolutionTotal)
	prometheus.MustRegister(DNSResolutionSuccess)
	prometheus.MustRegister(DNSResolutionFailure)
	prometheus.MustRegister(DNSResolutionDuration)
	prometheus.MustRegister(DNSResolutionConsistency)
	prometheus.MustRegister(DNSResponseSize)
	prometheus.MustRegister(DNSRecordCount)
	prometheus.MustRegister(DNSResolutionLatency)
	prometheus.MustRegister(DNSResolutionTTL)
	prometheus.MustRegister(DNSResolutionRetries)
	prometheus.MustRegister(DNSResolutionTimeout)
	prometheus.MustRegister(DNSResolutionNXDOMAIN)
	prometheus.MustRegister(DNSResolutionSERVFAIL)
	prometheus.MustRegister(DNSResolutionRefused)
	prometheus.MustRegister(DNSResolutionRateLimit)
	prometheus.MustRegister(DNSResolutionNetworkError)
	prometheus.MustRegister(DNSResolutionDNSSEC)
	prometheus.MustRegister(DNSResolutionEDNS)
	prometheus.MustRegister(DNSResolutionDNSSECSupport)
	prometheus.MustRegister(DNSResolutionProtocol)
	prometheus.MustRegister(DNSResolutionCacheHit)
	prometheus.MustRegister(DNSResolutionCacheMiss)

	// Register circuit breaker metrics
	prometheus.MustRegister(CircuitBreakerState)
	prometheus.MustRegister(CircuitBreakerFailures)

	// Register cache metrics
	prometheus.MustRegister(CacheSize)
	prometheus.MustRegister(CacheHits)
	prometheus.MustRegister(CacheMisses)
	prometheus.MustRegister(CacheEvictions)

	// Register health check metrics
	prometheus.MustRegister(HealthStatus)
	prometheus.MustRegister(HealthCheckDuration)
}
