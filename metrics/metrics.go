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

	DNSResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dns_response_size_bytes",
			Help:    "Size of DNS responses in bytes",
			Buckets: prometheus.ExponentialBuckets(64, 2, 10),
		},
		[]string{"server", "hostname"},
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
)
