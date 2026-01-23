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

func TestShardedCacheStats(t *testing.T) {
	cache := NewShardedCache(1024, 2)

	responseA := &dnsanalysis.DNSResponse{
		Server:    "8.8.8.8:53",
		Hostname:  "a.example.com",
		Addresses: []string{"1.1.1.1"},
	}
	responseB := &dnsanalysis.DNSResponse{
		Server:    "1.1.1.1:53",
		Hostname:  "b.example.com",
		Addresses: []string{"2.2.2.2"},
	}

	cache.Set("a.example.com", responseA, time.Minute)
	cache.Set("b.example.com", responseB, time.Minute)

	stats := cache.GetStats()
	if stats["entries"].(int) != 2 {
		t.Fatalf("expected 2 entries, got %v", stats["entries"])
	}
	if stats["size"].(int64) <= 0 {
		t.Fatalf("expected positive cache size, got %v", stats["size"])
	}
	if stats["max_size"].(int64) != 1024 {
		t.Fatalf("expected max size 1024, got %v", stats["max_size"])
	}
	if stats["num_shards"].(int) != 2 {
		t.Fatalf("expected num shards 2, got %v", stats["num_shards"])
	}
}

func TestShardedCacheEvictsOldest(t *testing.T) {
	cache := NewShardedCache(50, 1)

	responseA := &dnsanalysis.DNSResponse{
		Server:    "8.8.8.8:53",
		Hostname:  "old.example.com",
		Addresses: []string{"1.1.1.1"},
	}
	responseB := &dnsanalysis.DNSResponse{
		Server:    "1.1.1.1:53",
		Hostname:  "new.example.com",
		Addresses: []string{"2.2.2.2"},
	}

	cache.Set("old.example.com", responseA, 5*time.Second)
	cache.Set("new.example.com", responseB, 10*time.Second)

	if _, ok := cache.Get("old.example.com"); ok {
		t.Fatalf("expected oldest entry evicted")
	}
	if _, ok := cache.Get("new.example.com"); !ok {
		t.Fatalf("expected newest entry retained")
	}
}
