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

// DNSCache implements a simple DNS response cache
type DNSCache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	maxSize  int64
	cleanup  *time.Ticker
	stopChan chan struct{}
}

// New creates a new DNS cache
func New(maxSize int64) *DNSCache {
	cache := &DNSCache{
		entries:  make(map[string]*CacheEntry),
		maxSize:  maxSize,
		cleanup:  time.NewTicker(5 * time.Minute),
		stopChan: make(chan struct{}),
	}

	go cache.cleanupLoop()
	return cache
}

// Get retrieves a cached DNS response
func (c *DNSCache) Get(key string) (*dnsanalysis.DNSResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		cacheMisses.Inc()
		return nil, false
	}

	if time.Now().After(entry.Expires) {
		cacheMisses.Inc()
		delete(c.entries, key)
		return nil, false
	}

	cacheHits.Inc()
	return entry.Response, true
}

// Set stores a DNS response in the cache
func (c *DNSCache) Set(key string, response *dnsanalysis.DNSResponse, ttl time.Duration) {
	size := estimateSize(response)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove old entry if exists
	if _, ok := c.entries[key]; ok {
		delete(c.entries, key)
	}

	// Create new entry
	entry := &CacheEntry{
		Response: response,
		Expires:  time.Now().Add(ttl),
		Size:     size,
	}

	c.entries[key] = entry
	cacheSize.Set(float64(len(c.entries)))
}

// Delete removes a value from the cache
func (c *DNSCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.entries[key]; ok {
		delete(c.entries, key)
		cacheEvictions.Inc()
	}
}

// Clear removes all values from the cache
func (c *DNSCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	cacheSize.Set(0)
}

// Stop stops the cleanup goroutine
func (c *DNSCache) Stop() {
	c.cleanup.Stop()
	close(c.stopChan)
}

// cleanupLoop periodically removes expired entries
func (c *DNSCache) cleanupLoop() {
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
func (c *DNSCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.Expires) {
			delete(c.entries, key)
			cacheEvictions.Inc()
		}
	}
	cacheSize.Set(float64(len(c.entries)))
}

// GetStats returns cache statistics
func (c *DNSCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"entries":          len(c.entries),
		"hits":             cacheHits,
		"misses":           cacheMisses,
		"evictions":        cacheEvictions,
		"max_size":         c.maxSize,
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
