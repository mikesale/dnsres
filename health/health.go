package health

import (
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"dnsres/instrumentation"
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
	appLog  *log.Logger
	level   instrumentation.Level
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(servers []string, appLog *log.Logger, level instrumentation.Level) *HealthChecker {
	hc := &HealthChecker{
		servers: servers,
		status:  make(map[string]bool),
		appLog:  appLog,
		level:   level,
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
		if _, err := w.Write([]byte("healthy")); err != nil {
			log.Printf("health response write failed: %v", err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := w.Write([]byte("unhealthy")); err != nil {
			log.Printf("health response write failed: %v", err)
		}
	}
}

// checkLoop periodically checks the health of DNS servers
func (hc *HealthChecker) checkLoop() {
	hc.checkServers() // Run initial check immediately
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
		// Assume port 53 if not specified
		if !strings.Contains(server, ":") {
			server = server + ":53"
		}
		start := time.Now()
		// Simple TCP connection check
		conn, err := net.DialTimeout("tcp", server, 5*time.Second)
		if err != nil {
			hc.logf(instrumentation.Medium, "health check failed server=%s err=%v", server, err)
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

func (hc *HealthChecker) logf(level instrumentation.Level, format string, args ...any) {
	if hc.appLog == nil || hc.level < level {
		return
	}
	hc.appLog.Printf(format, args...)
}
