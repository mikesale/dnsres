package cache

import (
	"sync"
	"time"

	"dnsres/dnsanalysis"
	"dnsres/metrics"
)

// ShardedCache implements a sharded cache for DNS responses
type ShardedCache struct {
	shards    []*CacheShard
	numShards int
	maxSize   int64
}

// CacheShard represents a single shard in the cache
type CacheShard struct {
	entries map[string]*CacheEntry
	size    int64
	mu      sync.RWMutex
}

// CacheEntry represents a cached DNS response
type CacheEntry struct {
	Response *dnsanalysis.DNSResponse
	Expires  time.Time
	Size     int64
}

// NewShardedCache creates a new sharded cache
func NewShardedCache(maxSize int64, numShards int) *ShardedCache {
	if numShards <= 0 {
		numShards = 16 // Default number of shards
	}

	cache := &ShardedCache{
		shards:    make([]*CacheShard, numShards),
		numShards: numShards,
		maxSize:   maxSize,
	}

	for i := range cache.shards {
		cache.shards[i] = &CacheShard{
			entries: make(map[string]*CacheEntry),
		}
	}

	return cache
}

// Get retrieves a value from the cache
func (c *ShardedCache) Get(key string) (*dnsanalysis.DNSResponse, bool) {
	shard := c.getShard(key)
	shard.mu.RLock()

	entry, ok := shard.entries[key]
	if !ok {
		shard.mu.RUnlock()
		metrics.CacheMisses.Inc()
		return nil, false
	}

	if time.Now().After(entry.Expires) {
		shard.mu.RUnlock()

		shard.mu.Lock()
		// Double check under write lock
		if entry, ok := shard.entries[key]; ok && time.Now().After(entry.Expires) {
			delete(shard.entries, key)
			shard.size -= entry.Size
			metrics.CacheMisses.Inc()
		}
		shard.mu.Unlock()

		return nil, false
	}

	metrics.CacheHits.Inc()
	shard.mu.RUnlock()
	return entry.Response, true
}

// Set stores a value in the cache
func (c *ShardedCache) Set(key string, response *dnsanalysis.DNSResponse, ttl time.Duration) {
	shard := c.getShard(key)
	shard.mu.Lock()

	// Calculate entry size
	size := estimateSize(response)

	// Check if we need to evict entries
	if shard.size+size > c.maxSize/int64(c.numShards) {
		c.evictOldest(shard)
	}

	// Create new entry
	entry := &CacheEntry{
		Response: response,
		Expires:  time.Now().Add(ttl),
		Size:     size,
	}

	// Remove old entry if exists
	if old, ok := shard.entries[key]; ok {
		shard.size -= old.Size
	}

	// Add new entry
	shard.entries[key] = entry
	shard.size += size
	shard.mu.Unlock()

	metrics.CacheSize.Set(float64(c.getTotalEntries()))
}

// Delete removes a value from the cache
func (c *ShardedCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()

	if entry, ok := shard.entries[key]; ok {
		shard.size -= entry.Size
		delete(shard.entries, key)
		metrics.CacheEvictions.Inc()
		shard.mu.Unlock()
		metrics.CacheSize.Set(float64(c.getTotalEntries()))
		return
	}
	shard.mu.Unlock()
}

// Clear removes all values from the cache
func (c *ShardedCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.entries = make(map[string]*CacheEntry)
		shard.size = 0
		shard.mu.Unlock()
	}
	metrics.CacheSize.Set(0)
}

// GetStats returns cache statistics
func (c *ShardedCache) GetStats() map[string]interface{} {
	totalEntries, totalSize := c.getTotalStats()
	return map[string]interface{}{
		"entries":    totalEntries,
		"size":       totalSize,
		"hits":       float64(0), // Prometheus metrics are collected separately
		"misses":     float64(0), // Prometheus metrics are collected separately
		"evictions":  float64(0), // Prometheus metrics are collected separately
		"max_size":   c.maxSize,
		"num_shards": c.numShards,
	}
}

// getTotalStats returns the total number of entries and size across all shards
func (c *ShardedCache) getTotalStats() (int, int64) {
	var totalEntries int
	var totalSize int64

	for _, shard := range c.shards {
		shard.mu.RLock()
		totalEntries += len(shard.entries)
		totalSize += shard.size
		shard.mu.RUnlock()
	}

	return totalEntries, totalSize
}

// getTotalEntries returns the total number of entries across all shards
func (c *ShardedCache) getTotalEntries() int {
	entries, _ := c.getTotalStats()
	return entries
}

// getShard returns the shard for a given key
func (c *ShardedCache) getShard(key string) *CacheShard {
	hash := 0
	for _, b := range []byte(key) {
		hash = hash*31 + int(b)
	}
	// Ensure positive hash
	if hash < 0 {
		hash = -hash
	}
	return c.shards[hash%c.numShards]
}

// evictOldest removes the oldest entry from a shard
func (c *ShardedCache) evictOldest(shard *CacheShard) {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range shard.entries {
		if oldestKey == "" || entry.Expires.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Expires
		}
	}

	if oldestKey != "" {
		shard.size -= shard.entries[oldestKey].Size
		delete(shard.entries, oldestKey)
		metrics.CacheEvictions.Inc()
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
