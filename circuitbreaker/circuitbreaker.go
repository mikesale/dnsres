package circuitbreaker

import (
	"sync"
	"time"

	"dnsres/metrics"
)

// State represents the circuit breaker state
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	threshold int
	timeout   time.Duration
	failures  int
	lastError time.Time
	mu        sync.Mutex
	server    string // Add server field for metrics
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration, server string) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		server:    server,
	}
}

// Allow checks if the circuit breaker allows the request
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.failures >= cb.threshold {
		if time.Since(cb.lastError) < cb.timeout {
			metrics.CircuitBreakerState.WithLabelValues(cb.server).Set(float64(Open))
			return false
		}
		metrics.CircuitBreakerState.WithLabelValues(cb.server).Set(float64(HalfOpen))
	} else {
		metrics.CircuitBreakerState.WithLabelValues(cb.server).Set(float64(Closed))
	}

	metrics.CircuitBreakerFailures.WithLabelValues(cb.server).Inc()
	return true
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	metrics.CircuitBreakerState.WithLabelValues(cb.server).Set(float64(Closed))
	metrics.CircuitBreakerFailures.WithLabelValues(cb.server).Inc()
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastError = time.Now()
	metrics.CircuitBreakerFailures.WithLabelValues(cb.server).Inc()

	if cb.failures >= cb.threshold {
		metrics.CircuitBreakerState.WithLabelValues(cb.server).Set(float64(Open))
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.failures >= cb.threshold {
		if time.Since(cb.lastError) < cb.timeout {
			return "open"
		}
		return "half-open"
	}
	return "closed"
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	if !cb.Allow() {
		return nil, ErrCircuitOpen
	}

	result, err := fn()
	if err != nil {
		cb.RecordFailure()
		return nil, err
	}

	cb.RecordSuccess()
	return result, nil
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failures
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
}
