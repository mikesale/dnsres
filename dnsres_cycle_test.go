package dnsres

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"dnsres/circuitbreaker"
	"dnsres/dnsanalysis"
	"dnsres/metrics"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestResolveAllUpdatesCycleMetrics(t *testing.T) {
	hostname := "cycle.example.com"
	servers := []string{"1.1.1.1:53", "2.2.2.2:53"}

	breakers := make(map[string]*circuitbreaker.CircuitBreaker)
	stats := make(map[string]*ServerStats)
	for _, server := range servers {
		breakers[server] = circuitbreaker.NewCircuitBreaker(2, time.Minute, server)
		stats[server] = &ServerStats{}
	}

	resolver := &DNSResolver{
		config:     &Config{Hostnames: []string{hostname}, DNSServers: servers},
		breakers:   breakers,
		successLog: log.New(io.Discard, "", 0),
		errorLog:   log.New(io.Discard, "", 0),
		stats:      &ResolutionStats{Stats: stats, StartTime: time.Now()},
		resolveWithServerFunc: func(_ context.Context, server, host string) (*dnsanalysis.DNSResponse, error) {
			return &dnsanalysis.DNSResponse{
				Server:    server,
				Hostname:  host,
				Addresses: []string{"10.0.0.1"},
				TTL:       60,
			}, nil
		},
	}

	before := getHistogramCount(t, "dns_resolution_cycle_duration_seconds")
	resolver.resolveAll(context.Background())
	after := getHistogramCount(t, "dns_resolution_cycle_duration_seconds")

	if after != before+1 {
		t.Fatalf("expected cycle histogram count increment, got %d -> %d", before, after)
	}
	if got := testutil.ToFloat64(metrics.DNSResolutionConsistency.WithLabelValues(hostname)); got != 1 {
		t.Fatalf("expected consistency gauge 1, got %v", got)
	}
}

func getHistogramCount(t *testing.T, name string) uint64 {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		if len(mf.Metric) == 0 || mf.Metric[0].Histogram == nil {
			t.Fatalf("expected histogram metric for %s", name)
		}
		return mf.Metric[0].Histogram.GetSampleCount()
	}
	return 0
}
