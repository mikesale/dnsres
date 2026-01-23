package dnsres

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunLoopTicksAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticks := make(chan time.Time, 3)
	var calls int32

	resolver := &DNSResolver{
		config: &Config{QueryInterval: Duration{Duration: time.Second}},
		resolveAllFunc: func(context.Context) {
			atomic.AddInt32(&calls, 1)
		},
	}

	done := make(chan struct{})
	go func() {
		if err := resolver.runLoop(ctx, ticks); err != nil {
			t.Errorf("runLoop returned error: %v", err)
		}
		close(done)
	}()

	ticks <- time.Now()
	ticks <- time.Now()

	deadline := time.After(100 * time.Millisecond)
	for {
		if atomic.LoadInt32(&calls) == 2 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("expected 2 calls, got %d", atomic.LoadInt32(&calls))
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected runLoop to stop after cancel")
	}
}
