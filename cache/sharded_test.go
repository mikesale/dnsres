package cache

import (
	"testing"
	"time"

	"dnsres/dnsanalysis"
)

func TestShardedCacheTTLExpiration(t *testing.T) {
	cache := NewShardedCache(1024, 2)
	response := &dnsanalysis.DNSResponse{Hostname: "example.com"}
	cache.Set("example.com", response, 10*time.Millisecond)

	if _, ok := cache.Get("example.com"); !ok {
		t.Fatalf("expected cache hit before expiry")
	}

	time.Sleep(20 * time.Millisecond)
	if _, ok := cache.Get("example.com"); ok {
		t.Fatalf("expected cache miss after expiry")
	}
}

func TestShardedCacheEviction(t *testing.T) {
	cache := NewShardedCache(20, 1)

	responseA := &dnsanalysis.DNSResponse{
		Hostname:  "a.example.com",
		Addresses: []string{"1.1.1.1"},
	}
	responseB := &dnsanalysis.DNSResponse{
		Hostname:  "b.example.com",
		Addresses: []string{"2.2.2.2"},
	}

	cache.Set("a.example.com", responseA, time.Minute)
	cache.Set("b.example.com", responseB, time.Minute)

	entries, _ := cache.getTotalStats()
	if entries != 1 {
		t.Fatalf("expected eviction to keep 1 entry, got %d", entries)
	}
}
