package health

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dnsres/instrumentation"
	"dnsres/metrics"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestHealthCheckerTCPProbe(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listener: %v", err)
	}
	defer listener.Close()

	goodAddr := listener.Addr().String()
	badAddr := "127.0.0.1:1"

	hc := NewHealthChecker([]string{goodAddr, badAddr}, nil, instrumentation.None)
	hc.checkServers()

	hc.mu.RLock()
	goodStatus := hc.status[goodAddr]
	badStatus := hc.status[badAddr]
	hc.mu.RUnlock()

	if !goodStatus {
		t.Fatalf("expected healthy status for %s", goodAddr)
	}
	if badStatus {
		t.Fatalf("expected unhealthy status for %s", badAddr)
	}
}

func TestHealthCheckerHandler(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	hc := NewHealthChecker([]string{addr}, nil, instrumentation.None)
	hc.checkServers()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	hc.ServeHTTP(response, request)

	result := response.Result()
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed reading response: %v", err)
	}

	if result.StatusCode != http.StatusOK {
		t.Fatalf("expected status OK, got %d", result.StatusCode)
	}
	if string(body) != "healthy" {
		t.Fatalf("expected body healthy, got %s", string(body))
	}

	bad := NewHealthChecker([]string{"127.0.0.1:1"}, nil, instrumentation.None)
	bad.checkServers()
	badResponse := httptest.NewRecorder()
	bad.ServeHTTP(badResponse, request)
	if badResponse.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status unavailable, got %d", badResponse.Code)
	}
}

func TestHealthCheckerLoopStarts(t *testing.T) {
	hc := NewHealthChecker([]string{"127.0.0.1:1"}, nil, instrumentation.None)

	select {
	case <-time.After(10 * time.Millisecond):
		if hc == nil {
			t.Fatalf("unexpected nil health checker")
		}
	}
}

func TestHealthCheckerMetrics(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to open listener: %v", err)
	}
	defer listener.Close()

	goodAddr := listener.Addr().String()
	badAddr := "127.0.0.1:1"

	beforeSuccess := testutil.ToFloat64(metrics.DNSResolutionSuccess.WithLabelValues(goodAddr, ""))
	beforeFailure := testutil.ToFloat64(metrics.DNSResolutionFailure.WithLabelValues(badAddr, "", "health_check"))

	hc := NewHealthChecker([]string{goodAddr, badAddr}, nil, instrumentation.None)
	hc.checkServers()

	afterSuccess := testutil.ToFloat64(metrics.DNSResolutionSuccess.WithLabelValues(goodAddr, ""))
	afterFailure := testutil.ToFloat64(metrics.DNSResolutionFailure.WithLabelValues(badAddr, "", "health_check"))

	if afterSuccess <= beforeSuccess {
		t.Fatalf("expected success metric to increment")
	}
	if afterFailure <= beforeFailure {
		t.Fatalf("expected failure metric to increment")
	}
}
