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
