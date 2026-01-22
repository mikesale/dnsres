package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DNS Resolution Metrics
	DNSResolutionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_total",
			Help: "Total number of DNS resolution attempts",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_success",
			Help: "Number of successful DNS resolutions",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionFailure = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_failure",
			Help: "Number of failed DNS resolutions",
		},
		[]string{"server", "hostname", "error_type"},
	)

	DNSResolutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_duration_seconds",
			Help:    "DNS resolution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionConsistency = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_consistency",
			Help: "Whether DNS responses are consistent across servers",
		},
		[]string{"hostname"},
	)

	DNSResolutionCycleDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_cycle_duration_seconds",
			Help:    "Duration of a full resolution cycle in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	DNSResolutionTTL = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_resolution_ttl_seconds",
			Help:    "TTL values from DNS responses",
			Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
		},
		[]string{"server", "hostname", "record_type"},
	)

	DNSResolutionRetries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_retries_total",
			Help: "Total number of DNS resolution retry attempts",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionTimeout = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_timeout_total",
			Help: "Total number of DNS resolution timeouts",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionNXDOMAIN = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_nxdomain_total",
			Help: "Total number of NXDOMAIN responses",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionSERVFAIL = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_servfail_total",
			Help: "Total number of SERVFAIL responses",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionRefused = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_refused_total",
			Help: "Total number of REFUSED responses",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionRateLimit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_rate_limit_total",
			Help: "Total number of rate limit occurrences",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionNetworkError = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_network_error_total",
			Help: "Total number of network-related errors",
		},
		[]string{"server", "hostname", "error_type"},
	)

	DNSResolutionDNSSEC = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_dnssec_total",
			Help: "Total number of DNSSEC validation results",
		},
		[]string{"server", "hostname", "status"},
	)

	DNSResolutionEDNS = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_edns_support",
			Help: "EDNS support status (1=supported, 0=not supported)",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionDNSSECSupport = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "dns_resolution_dnssec_support",
			Help: "DNSSEC support status (1=supported, 0=not supported)",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionProtocol = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_protocol_total",
			Help: "Total number of DNS resolutions by protocol",
		},
		[]string{"server", "hostname", "protocol"},
	)

	DNSResolutionCacheHit = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_cache_hit",
			Help: "Number of cache hits",
		},
		[]string{"server", "hostname"},
	)

	DNSResolutionCacheMiss = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dns_resolution_cache_miss",
			Help: "Number of cache misses",
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

	// Circuit Breaker Metrics
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breaker (0=Closed, 1=Open, 2=Half-Open)",
		},
		[]string{"server"},
	)

	CircuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures",
			Help: "Number of consecutive failures for each server",
		},
		[]string{"server"},
	)

	// Health Check Metrics
	HealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health_status",
			Help: "Health status of each DNS server (1=Healthy, 0=Unhealthy)",
		},
		[]string{"server"},
	)

	HealthCheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "health_check_duration_seconds",
			Help:    "Duration of health checks in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"server"},
	)

	DNSRecordCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_record_count",
			Help:    "Number of records in DNS responses",
			Buckets: prometheus.LinearBuckets(0, 1, 20),
		},
		[]string{"server", "hostname", "type"},
	)

	DNSResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_response_size_bytes",
			Help:    "Size of DNS responses in bytes",
			Buckets: prometheus.ExponentialBuckets(64, 2, 10),
		},
		[]string{"server", "hostname"},
	)
)
