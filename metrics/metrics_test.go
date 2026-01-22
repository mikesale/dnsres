package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestDNSResolutionCycleDurationMetric(t *testing.T) {
	registry := prometheus.NewRegistry()
	registry.MustRegister(DNSResolutionCycleDuration)

	before := getHistogramCount(t, registry, "dns_resolution_cycle_duration_seconds")
	DNSResolutionCycleDuration.Observe(0.123)
	after := getHistogramCount(t, registry, "dns_resolution_cycle_duration_seconds")

	if after != before+1 {
		t.Fatalf("expected histogram count to increment by 1, got %d -> %d", before, after)
	}
}

func getHistogramCount(t *testing.T, registry *prometheus.Registry, name string) uint64 {
	mfs, err := registry.Gather()
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
