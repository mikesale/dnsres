package dnsanalysis

import (
	"testing"
)

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
