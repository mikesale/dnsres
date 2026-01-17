package dnspool

import (
	"strings"
	"sync"
	"time"

	"dnsres/metrics"

	"github.com/miekg/dns"
)

// ClientPool manages a pool of DNS clients
type ClientPool struct {
	clients map[string][]*dns.Client
	mu      sync.Mutex
	MaxSize int
	Timeout time.Duration
}

// NewClientPool creates a new DNS client pool
func NewClientPool(maxSize int, timeout time.Duration) *ClientPool {
	return &ClientPool{
		clients: make(map[string][]*dns.Client),
		MaxSize: maxSize,
		Timeout: timeout,
	}
}

// Get retrieves a client from the pool or creates a new one
func (p *ClientPool) Get(server string) (*dns.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Assume port 53 if not specified
	if !strings.Contains(server, ":") {
		server = server + ":53"
	}

	clients := p.clients[server]
	if len(clients) > 0 {
		client := clients[len(clients)-1]
		p.clients[server] = clients[:len(clients)-1]
		metrics.DNSResolutionProtocol.WithLabelValues(server, "", "pooled").Inc()
		return client, nil
	}

	client := &dns.Client{
		Timeout: p.Timeout,
	}
	metrics.DNSResolutionProtocol.WithLabelValues(server, "", "new").Inc()
	return client, nil
}

// Put returns a client to the pool
func (p *ClientPool) Put(server string, client *dns.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reset client state
	client.Timeout = p.Timeout

	// Add to pool if not at max size
	if len(p.clients[server]) < p.MaxSize {
		p.clients[server] = append(p.clients[server], client)
		metrics.DNSResolutionProtocol.WithLabelValues(server, "", "returned").Inc()
	} else {
		metrics.DNSResolutionProtocol.WithLabelValues(server, "", "dropped").Inc()
	}
}

// GetStats returns pool statistics
func (p *ClientPool) GetStats() map[string]interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	totalClients := 0
	for _, clients := range p.clients {
		totalClients += len(clients)
	}

	return map[string]interface{}{
		"total_clients": totalClients,
		"max_size":      p.MaxSize,
		"timeout":       p.Timeout.String(),
	}
}
