package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// ResolutionStats tracks resolution statistics
type ResolutionStats struct {
	Total     int
	Failures  int
	LastError string
	StartTime time.Time
	Stats     map[string]*ServerStats
}

// ServerStats tracks statistics for a single server
type ServerStats struct {
	Total     int
	Failures  int
	LastError string
}

// GenerateReport generates a statistics report
func (r *DNSResolver) GenerateReport() string {
	var report strings.Builder
	report.WriteString("Hour              | DNS Server     | Total    | Fails    | Fail %  \n")
	report.WriteString("-----------------------------------------------------------------\n")

	// Sort stats by time
	keys := make([]string, 0, len(r.stats.Stats))
	for k := range r.stats.Stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		stats := r.stats.Stats[key]
		server := key
		hour := r.stats.StartTime.Format("2006-01-02 15:04")
		failPercent := 0.0
		if stats.Total > 0 {
			failPercent = float64(stats.Failures) / float64(stats.Total) * 100
		}
		report.WriteString(fmt.Sprintf("%s | %-12s | %-8d | %-8d | %6.2f%%\n",
			hour, server, stats.Total, stats.Failures, failPercent))
	}

	return report.String()
}
