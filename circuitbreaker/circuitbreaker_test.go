package circuitbreaker

import (
	"testing"
	"time"
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
