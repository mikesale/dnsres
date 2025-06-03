package cache

import (
	"sync"
	"time"

	"dnsres/dnsanalysis"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheSize      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "dns_cache_size", Help: "Current size of DNS cache"})
	cacheHits      = prometheus.NewCounter(prometheus.CounterOpts{Name: "dns_cache_hits_total", Help: "Total number of cache hits"})
	cacheMisses    = prometheus.NewCounter(prometheus.CounterOpts{Name: "dns_cache_misses_total", Help: "Total number of cache misses"})
	cacheEvictions = prometheus.NewCounter(prometheus.CounterOpts{Name: "dns_cache_evictions_total", Help: "Total number of cache evictions"})
)

// CacheEntry represents a cached DNS response
type CacheEntry struct {
	Response *dnsanalysis.DNSResponse
	Expires  time.Time
	Size     int64
}

// Shard represents a single shard in the cache
type Shard struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
}

// ShardedCache implements a sharded DNS response cache
type ShardedCache struct {
	shards    []*Shard
	maxSize   int64
	cleanup   *time.Ticker
	stopChan  chan struct{}
	numShards int
}

// NewShardedCache creates a new sharded DNS cache
func NewShardedCache(maxSize int64, numShards int) *ShardedCache {
	if numShards <= 0 {
		numShards = 16 // Default number of shards
	}

	cache := &ShardedCache{
		shards:    make([]*Shard, numShards),
		maxSize:   maxSize,
		cleanup:   time.NewTicker(5 * time.Minute),
		stopChan:  make(chan struct{}),
		numShards: numShards,
	}

	// Initialize shards
	for i := 0; i < numShards; i++ {
		cache.shards[i] = &Shard{
			entries: make(map[string]*CacheEntry),
		}
	}

	go cache.cleanupLoop()
	return cache
}

// getShard returns the shard for a given key
func (c *ShardedCache) getShard(key string) *Shard {
	hash := 0
	for _, ch := range key {
		hash = 31*hash + int(ch)
	}
	return c.shards[abs(hash)%c.numShards]
}

// Get retrieves a cached DNS response
func (c *ShardedCache) Get(key string) (*dnsanalysis.DNSResponse, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	entry, ok := shard.entries[key]
	if !ok {
		cacheMisses.Inc()
		return nil, false
	}

	if time.Now().After(entry.Expires) {
		cacheMisses.Inc()
		delete(shard.entries, key)
		return nil, false
	}

	cacheHits.Inc()
	return entry.Response, true
}

// Set stores a DNS response in the cache
func (c *ShardedCache) Set(key string, response *dnsanalysis.DNSResponse, ttl time.Duration) {
	size := estimateSize(response)
	shard := c.getShard(key)

	shard.mu.Lock()
	defer shard.mu.Unlock()

	// Remove old entry if exists
	if _, ok := shard.entries[key]; ok {
		delete(shard.entries, key)
	}

	// Create new entry
	entry := &CacheEntry{
		Response: response,
		Expires:  time.Now().Add(ttl),
		Size:     size,
	}

	shard.entries[key] = entry
	cacheSize.Set(float64(c.getTotalEntries()))
}

// Delete removes a value from the cache
func (c *ShardedCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, ok := shard.entries[key]; ok {
		delete(shard.entries, key)
		cacheEvictions.Inc()
	}
}

// Clear removes all values from the cache
func (c *ShardedCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.entries = make(map[string]*CacheEntry)
		shard.mu.Unlock()
	}
	cacheSize.Set(0)
}

// Stop stops the cleanup goroutine
func (c *ShardedCache) Stop() {
	c.cleanup.Stop()
	close(c.stopChan)
}

// cleanupLoop periodically removes expired entries
func (c *ShardedCache) cleanupLoop() {
	for {
		select {
		case <-c.cleanup.C:
			c.cleanupExpired()
		case <-c.stopChan:
			return
		}
	}
}

// cleanupExpired removes expired entries from the cache
func (c *ShardedCache) cleanupExpired() {
	now := time.Now()
	for _, shard := range c.shards {
		shard.mu.Lock()
		for key, entry := range shard.entries {
			if now.After(entry.Expires) {
				delete(shard.entries, key)
				cacheEvictions.Inc()
			}
		}
		shard.mu.Unlock()
	}
	cacheSize.Set(float64(c.getTotalEntries()))
}

// getTotalEntries returns the total number of entries across all shards
func (c *ShardedCache) getTotalEntries() int {
	total := 0
	for _, shard := range c.shards {
		shard.mu.RLock()
		total += len(shard.entries)
		shard.mu.RUnlock()
	}
	return total
}

// GetStats returns cache statistics
func (c *ShardedCache) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"entries":          c.getTotalEntries(),
		"hits":             float64(0), // Prometheus metrics are collected separately
		"misses":           float64(0), // Prometheus metrics are collected separately
		"evictions":        float64(0), // Prometheus metrics are collected separately
		"max_size":         c.maxSize,
		"num_shards":       c.numShards,
		"cleanup_interval": "5m",
	}
}

// estimateSize estimates the size of a DNS response
func estimateSize(response *dnsanalysis.DNSResponse) int64 {
	size := int64(len(response.Server) + len(response.Hostname))
	for _, addr := range response.Addresses {
		size += int64(len(addr))
	}
	return size
}

// abs returns the absolute value of an integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
