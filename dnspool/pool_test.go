package dnspool

import (
	"testing"
	"time"
)

func TestClientPoolReuse(t *testing.T) {
	pool := NewClientPool(2, 2*time.Second)

	client, err := pool.Get("8.8.8.8:53")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	client.Timeout = 500 * time.Millisecond
	pool.Put("8.8.8.8:53", client)

	reused, err := pool.Get("8.8.8.8:53")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reused != client {
		t.Fatalf("expected pooled client reuse")
	}
	if reused.Timeout != 2*time.Second {
		t.Fatalf("expected timeout reset, got %s", reused.Timeout)
	}
}

func TestClientPoolMaxSize(t *testing.T) {
	pool := NewClientPool(1, time.Second)

	first, err := pool.Get("8.8.8.8:53")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := pool.Get("8.8.8.8:53")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pool.Put("8.8.8.8:53", first)
	pool.Put("8.8.8.8:53", second)

	stats := pool.GetStats()
	if stats["total_clients"].(int) != 1 {
		t.Fatalf("expected pool size 1, got %v", stats["total_clients"])
	}
}
