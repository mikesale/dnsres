package circuitbreaker

import (
	"testing"
	"time"

	"dnsres/metrics"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCircuitBreakerStateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(2, 20*time.Millisecond, "server")

	if state := cb.GetState(); state != "closed" {
		t.Fatalf("expected initial state closed, got %s", state)
	}

	cb.RecordFailure()
	if state := cb.GetState(); state != "closed" {
		t.Fatalf("expected state closed after 1 failure, got %s", state)
	}

	cb.RecordFailure()
	if state := cb.GetState(); state != "open" {
		t.Fatalf("expected state open after threshold, got %s", state)
	}

	if allowed := cb.Allow(); allowed {
		t.Fatalf("expected Allow to return false while open")
	}

	time.Sleep(25 * time.Millisecond)
	if allowed := cb.Allow(); !allowed {
		t.Fatalf("expected Allow to return true after timeout")
	}
	if state := cb.GetState(); state != "half-open" {
		t.Fatalf("expected state half-open after timeout, got %s", state)
	}

	cb.RecordSuccess()
	if state := cb.GetState(); state != "closed" {
		t.Fatalf("expected state closed after success, got %s", state)
	}
}

func TestCircuitBreakerFailureMetricOnlyOnFailure(t *testing.T) {
	server := "metric-test-server"
	cb := NewCircuitBreaker(5, time.Second, server)

	before := testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server))

	// Allow should not increment failure counter
	cb.Allow()
	after := testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server))
	if after != before {
		t.Errorf("Allow() incremented failure counter: before=%v after=%v", before, after)
	}

	// RecordSuccess should not increment failure counter
	cb.RecordSuccess()
	after = testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server))
	if after != before {
		t.Errorf("RecordSuccess() incremented failure counter: before=%v after=%v", before, after)
	}

	// RecordFailure should increment failure counter
	cb.RecordFailure()
	after = testutil.ToFloat64(metrics.CircuitBreakerFailures.WithLabelValues(server))
	if after != before+1 {
		t.Errorf("RecordFailure() did not increment failure counter: before=%v after=%v", before, after)
	}
}
