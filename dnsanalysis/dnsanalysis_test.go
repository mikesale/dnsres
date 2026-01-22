package dnsanalysis

import (
	"context"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestAnalyzeResponseMetrics(t *testing.T) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	msg.Answer = append(msg.Answer, &dns.A{
		Hdr: dns.RR_Header{
			Name:   dns.Fqdn("example.com"),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
		A: []byte{1, 2, 3, 4},
	})
	msg.Answer = append(msg.Answer, &dns.RRSIG{
		Hdr: dns.RR_Header{
			Name:   dns.Fqdn("example.com"),
			Rrtype: dns.TypeRRSIG,
			Class:  dns.ClassINET,
			Ttl:    300,
		},
	})
	msg.Extra = append(msg.Extra, &dns.OPT{Hdr: dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT}})

	analysis, err := AnalyzeResponse(context.Background(), "server", "example.com", msg, msg.Len(), "udp", 10*time.Millisecond)
	if err != nil {
		t.Fatalf("AnalyzeResponse returned error: %v", err)
	}
	if analysis.RecordCount["A"] != 1 {
		t.Fatalf("expected A record count 1, got %d", analysis.RecordCount["A"])
	}
	if analysis.TTL != 300 {
		t.Fatalf("expected TTL 300, got %d", analysis.TTL)
	}
	if !analysis.DNSSEC {
		t.Fatalf("expected DNSSEC true")
	}
	if !analysis.EDNS {
		t.Fatalf("expected EDNS true")
	}
}

func TestCompareResponses(t *testing.T) {
	base := &DNSResponse{
		Server:    "server-1",
		Hostname:  "example.com",
		Addresses: []string{"1.1.1.1", "2.2.2.2"},
	}
	match := &DNSResponse{
		Server:    "server-2",
		Hostname:  "example.com",
		Addresses: []string{"2.2.2.2", "1.1.1.1"},
	}
	mismatch := &DNSResponse{
		Server:    "server-3",
		Hostname:  "example.com",
		Addresses: []string{"9.9.9.9"},
	}

	if !CompareResponses([]*DNSResponse{base, match}) {
		t.Fatalf("expected responses to match")
	}
	if CompareResponses([]*DNSResponse{base, mismatch}) {
		t.Fatalf("expected responses to mismatch")
	}
}
