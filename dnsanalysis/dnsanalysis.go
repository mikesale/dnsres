package dnsanalysis

import (
	"time"

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
