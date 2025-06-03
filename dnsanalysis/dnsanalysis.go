package dnsanalysis

import (
	"context"
	"time"

	"dnsres/metrics"

	"github.com/miekg/dns"
)

// DNSResponse represents a DNS resolution response
type DNSResponse struct {
	Server      string
	Hostname    string
	Addresses   []string
	Response    *dns.Msg
	RecordCount map[string]int
	TTL         uint32
	Size        int
	DNSSEC      bool
	EDNS        bool
	Protocol    string
	Duration    time.Duration
}

// AnalyzeResponse analyzes a DNS response and updates metrics
func AnalyzeResponse(ctx context.Context, server, hostname string, response *dns.Msg, size int, protocol string, duration time.Duration) (*DNSResponse, error) {
	analysis := &DNSResponse{
		Server:      server,
		Hostname:    hostname,
		Response:    response,
		RecordCount: make(map[string]int),
		TTL:         0,
		Size:        size,
		Protocol:    protocol,
		Duration:    duration,
	}

	// Count records by type
	for _, rr := range response.Answer {
		recordType := dns.TypeToString[rr.Header().Rrtype]
		analysis.RecordCount[recordType]++
		analysis.TTL = rr.Header().Ttl

		// Update metrics
		metrics.DNSRecordCount.WithLabelValues(server, hostname, recordType).Observe(float64(analysis.RecordCount[recordType]))
		metrics.DNSResolutionTTL.WithLabelValues(server, hostname, recordType).Observe(float64(rr.Header().Ttl))
	}

	// Check for DNSSEC
	analysis.DNSSEC = hasDNSSEC(response)
	metrics.DNSResolutionDNSSECSupport.WithLabelValues(server, hostname).Set(boolToFloat64(analysis.DNSSEC))

	// Check for EDNS
	analysis.EDNS = hasEDNS(response)
	metrics.DNSResolutionEDNS.WithLabelValues(server, hostname).Set(boolToFloat64(analysis.EDNS))

	// Update response size metric
	metrics.DNSResponseSize.WithLabelValues(server, hostname).Observe(float64(size))

	// Update protocol metric
	metrics.DNSResolutionProtocol.WithLabelValues(server, hostname, protocol).Inc()

	return analysis, nil
}

// CompareResponses compares multiple DNS responses for consistency
func CompareResponses(responses []*DNSResponse) bool {
	if len(responses) <= 1 {
		return true
	}

	// Compare addresses from first response with others
	firstAddrs := make(map[string]struct{})
	for _, addr := range responses[0].Addresses {
		firstAddrs[addr] = struct{}{}
	}

	for i := 1; i < len(responses); i++ {
		if len(responses[i].Addresses) != len(responses[0].Addresses) {
			return false
		}

		for _, addr := range responses[i].Addresses {
			if _, ok := firstAddrs[addr]; !ok {
				return false
			}
		}
	}

	return true
}

// getMinTTL returns the minimum TTL from a DNS response
func getMinTTL(msg *dns.Msg) uint32 {
	if len(msg.Answer) == 0 {
		return 0
	}

	minTTL := msg.Answer[0].Header().Ttl
	for _, rr := range msg.Answer {
		if rr.Header().Ttl < minTTL {
			minTTL = rr.Header().Ttl
		}
	}
	return minTTL
}

// Helper functions
func hasDNSSEC(msg *dns.Msg) bool {
	for _, rr := range msg.Answer {
		if rr.Header().Rrtype == dns.TypeRRSIG || rr.Header().Rrtype == dns.TypeDNSKEY {
			return true
		}
	}
	return false
}

func hasEDNS(msg *dns.Msg) bool {
	return msg.IsEdns0() != nil
}

func compareRecordCounts(count1, count2 map[string]int) bool {
	if len(count1) != len(count2) {
		return false
	}
	for k, v := range count1 {
		if count2[k] != v {
			return false
		}
	}
	return true
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
