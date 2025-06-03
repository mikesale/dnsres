package health

import (
	"net"
	"net/http"
	"sync"
	"time"

	"dnsres/metrics"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Details   map[string]string `json:"details,omitempty"`
}

// HealthChecker implements a health check endpoint
type HealthChecker struct {
	servers []string
	status  map[string]bool
	mu      sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(servers []string) *HealthChecker {
	hc := &HealthChecker{
		servers: servers,
		status:  make(map[string]bool),
	}
	go hc.checkLoop()
	return hc
}

// ServeHTTP implements the http.Handler interface
func (hc *HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Check if any server is healthy
	healthy := false
	for _, status := range hc.status {
		if status {
			healthy = true
			break
		}
	}

	if healthy {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("unhealthy"))
	}
}

// checkLoop periodically checks the health of DNS servers
func (hc *HealthChecker) checkLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		hc.checkServers()
	}
}

// checkServers checks the health of all DNS servers
func (hc *HealthChecker) checkServers() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, server := range hc.servers {
		start := time.Now()
		// Simple TCP connection check
		conn, err := net.DialTimeout("tcp", server, 5*time.Second)
		if err != nil {
			hc.status[server] = false
			metrics.DNSResolutionFailure.WithLabelValues(server, "", "health_check").Inc()
			continue
		}
		conn.Close()
		hc.status[server] = true
		metrics.DNSResolutionSuccess.WithLabelValues(server, "").Inc()
		metrics.DNSResolutionDuration.WithLabelValues(server, "").Observe(time.Since(start).Seconds())
	}
}

// Helper functions
func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
