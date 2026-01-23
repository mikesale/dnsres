package circuitbreaker

import (
	"testing"
	"time"

	"dnsres/metrics"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCircuitBreakerMetrics(t *testing.T) {
	server := "metrics-server"
	cb := NewCircuitBreaker(1, 20*time.Millisecond, server)

	startFailures := testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server))
	cb.RecordFailure()
	if got := testutil.ToFloat64(metrics.CircuitBreakerState.WithLabelValues(server)); got != float64(Open) {
		t.Fatalf("expected open state metric, got %v", got)
	}
	if got := testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server)); got <= startFailures {
		t.Fatalf("expected failure metric increment")
	}

	if allowed := cb.Allow(); allowed {
		t.Fatalf("expected Allow false while open")
	}
	if got := testutil.ToFloat64(metrics.CircuitBreakerState.WithLabelValues(server)); got != float64(Open) {
		t.Fatalf("expected open state metric after allow, got %v", got)
	}

	time.Sleep(25 * time.Millisecond)
	if allowed := cb.Allow(); !allowed {
		t.Fatalf("expected Allow true after timeout")
	}
	if got := testutil.ToFloat64(metrics.CircuitBreakerState.WithLabelValues(server)); got != float64(HalfOpen) {
		t.Fatalf("expected half-open state metric, got %v", got)
	}

	cb.RecordSuccess()
	if got := testutil.ToFloat64(metrics.CircuitBreakerState.WithLabelValues(server)); got != float64(Closed) {
		t.Fatalf("expected closed state metric, got %v", got)
	}
}
