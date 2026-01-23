package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"dnsres/cache"
	"dnsres/circuitbreaker"
	"dnsres/metrics"

	"github.com/miekg/dns"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeDNSClient struct {
	response *dns.Msg
	err      error
}

func (c *fakeDNSClient) ExchangeContext(ctx context.Context, msg *dns.Msg, server string) (*dns.Msg, time.Duration, error) {
	return c.response, 0, c.err
}

func TestResolveWithServerCircuitBreakerOpen(t *testing.T) {
	server := "8.8.8.8:53"
	breaker := circuitbreaker.NewCircuitBreaker(1, time.Minute, server)
	breaker.RecordFailure()

	resolver := &DNSResolver{
		breakers: map[string]*circuitbreaker.CircuitBreaker{server: breaker},
		cache:    cache.NewShardedCache(1024, 1),
		getClient: func(string) (dnsClient, error) {
			t.Fatalf("unexpected client fetch")
			return nil, nil
		},
		putClient: func(string, dnsClient) {},
	}

	_, err := resolver.resolveWithServer(context.Background(), server, "example.com")
	if err == nil || !strings.Contains(err.Error(), "circuit breaker open") {
		t.Fatalf("expected circuit breaker error, got %v", err)
	}
}

func TestResolveWithServerClientPoolError(t *testing.T) {
	server := "8.8.8.8:53"
	resolver := &DNSResolver{
		breakers: map[string]*circuitbreaker.CircuitBreaker{
			server: circuitbreaker.NewCircuitBreaker(2, time.Minute, server),
		},
		cache: cache.NewShardedCache(1024, 1),
		getClient: func(string) (dnsClient, error) {
			return nil, errors.New("pool unavailable")
		},
		putClient: func(string, dnsClient) {},
	}

	_, err := resolver.resolveWithServer(context.Background(), server, "example.com")
	if err == nil || !strings.Contains(err.Error(), "failed to get client from pool") {
		t.Fatalf("expected client pool error, got %v", err)
	}
}

func TestResolveWithServerQueryError(t *testing.T) {
	server := "8.8.8.8:53"
	fake := &fakeDNSClient{err: errors.New("exchange failed")}

	resolver := &DNSResolver{
		breakers: map[string]*circuitbreaker.CircuitBreaker{
			server: circuitbreaker.NewCircuitBreaker(2, time.Minute, server),
		},
		cache: cache.NewShardedCache(1024, 1),
		stats: &ResolutionStats{Stats: map[string]*ServerStats{server: {}}},
		getClient: func(string) (dnsClient, error) {
			return fake, nil
		},
		putClient: func(string, dnsClient) {},
	}

	_, err := resolver.resolveWithServer(context.Background(), server, "example.com")
	if err == nil || !strings.Contains(err.Error(), "DNS query failed") {
		t.Fatalf("expected DNS query error, got %v", err)
	}
	if resolver.stats.Stats[server].Failures != 1 {
		t.Fatalf("expected failures incremented, got %d", resolver.stats.Stats[server].Failures)
	}
}

func TestResolveWithServerRcodeError(t *testing.T) {
	server := "8.8.8.8:53"
	response := new(dns.Msg)
	response.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	response.Rcode = dns.RcodeNameError

	fake := &fakeDNSClient{response: response}
	resolver := &DNSResolver{
		breakers: map[string]*circuitbreaker.CircuitBreaker{
			server: circuitbreaker.NewCircuitBreaker(2, time.Minute, server),
		},
		cache: cache.NewShardedCache(1024, 1),
		stats: &ResolutionStats{Stats: map[string]*ServerStats{server: {}}},
		getClient: func(string) (dnsClient, error) {
			return fake, nil
		},
		putClient: func(string, dnsClient) {},
	}

	_, err := resolver.resolveWithServer(context.Background(), server, "example.com")
	if err == nil || !strings.Contains(err.Error(), "NXDOMAIN") {
		t.Fatalf("expected NXDOMAIN error, got %v", err)
	}
	if resolver.stats.Stats[server].Failures != 1 {
		t.Fatalf("expected failures incremented, got %d", resolver.stats.Stats[server].Failures)
	}
}

func TestResolveWithServerSuccessUpdatesMetrics(t *testing.T) {
	server := "9.9.9.9:53"
	response := new(dns.Msg)
	response.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	response.Answer = append(response.Answer, &dns.A{
		Hdr: dns.RR_Header{
			Name:   dns.Fqdn("example.com"),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		A: []byte{1, 1, 1, 1},
	})

	fake := &fakeDNSClient{response: response}
	resolver := &DNSResolver{
		breakers: map[string]*circuitbreaker.CircuitBreaker{
			server: circuitbreaker.NewCircuitBreaker(2, time.Minute, server),
		},
		cache: cache.NewShardedCache(1024, 1),
		stats: &ResolutionStats{Stats: map[string]*ServerStats{server: {}}},
		getClient: func(string) (dnsClient, error) {
			return fake, nil
		},
		putClient: func(string, dnsClient) {},
	}

	beforeSuccess := testutil.ToFloat64(metrics.DNSResolutionSuccess.WithLabelValues(server, "example.com"))
	resp, err := resolver.resolveWithServer(context.Background(), server, "example.com")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(resp.Addresses) != 1 || resp.Addresses[0] != "1.1.1.1" {
		t.Fatalf("unexpected response addresses: %+v", resp.Addresses)
	}
	if resp.TTL != 300 {
		t.Fatalf("expected TTL 300, got %d", resp.TTL)
	}
	if resolver.stats.Stats[server].Total != 1 {
		t.Fatalf("expected total incremented, got %d", resolver.stats.Stats[server].Total)
	}
	if resolver.stats.Stats[server].Failures != 0 {
		t.Fatalf("expected no failures, got %d", resolver.stats.Stats[server].Failures)
	}
	if _, ok := resolver.cache.Get("example.com"); !ok {
		t.Fatalf("expected response cached")
	}
	afterSuccess := testutil.ToFloat64(metrics.DNSResolutionSuccess.WithLabelValues(server, "example.com"))
	if afterSuccess <= beforeSuccess {
		t.Fatalf("expected success metric increment")
	}
}
