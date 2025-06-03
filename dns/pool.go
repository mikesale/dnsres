package dnspool

import (
	"sync"
	"time"

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

	clients := p.clients[server]
	if len(clients) > 0 {
		client := clients[len(clients)-1]
		p.clients[server] = clients[:len(clients)-1]
		return client, nil
	}

	client := &dns.Client{
		Timeout: p.Timeout,
	}
	return client, nil
}

// Put returns a client to the pool
func (p *ClientPool) Put(client *dns.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Reset client state
	client.Timeout = p.Timeout

	// Add to pool if not at max size
	if len(p.clients[""]) < p.MaxSize {
		p.clients[""] = append(p.clients[""], client)
	}
}
