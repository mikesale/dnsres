package cache

import (
	"sync"
	"time"

	"dnsres/dnsanalysis"
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
	defer shard.mu.RUnlock()

	entry, ok := shard.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.Expires) {
		shard.mu.RUnlock()
		shard.mu.Lock()
		delete(shard.entries, key)
		shard.size -= entry.Size
		shard.mu.Unlock()
		shard.mu.RLock()
		return nil, false
	}

	return entry.Response, true
}

// Set stores a value in the cache
func (c *ShardedCache) Set(key string, response *dnsanalysis.DNSResponse, ttl time.Duration) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

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
}

// Delete removes a value from the cache
func (c *ShardedCache) Delete(key string) {
	shard := c.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	if entry, ok := shard.entries[key]; ok {
		shard.size -= entry.Size
		delete(shard.entries, key)
	}
}

// Clear removes all values from the cache
func (c *ShardedCache) Clear() {
	for _, shard := range c.shards {
		shard.mu.Lock()
		shard.entries = make(map[string]*CacheEntry)
		shard.size = 0
		shard.mu.Unlock()
	}
}

// GetStats returns cache statistics
func (c *ShardedCache) GetStats() (int, int64) {
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

// getShard returns the shard for a given key
func (c *ShardedCache) getShard(key string) *CacheShard {
	hash := 0
	for _, b := range []byte(key) {
		hash = hash*31 + int(b)
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
