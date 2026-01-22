package instrumentation

import (
	"fmt"
	"strings"
)

// Level controls how much instrumentation logging is emitted.
type Level int

const (
	None Level = iota
	Low
	Medium
	High
	Critical
)

func ParseLevel(value string) (Level, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return None, nil
	}

	switch normalized {
	case "none":
		return None, nil
	case "low":
		return Low, nil
	case "medium":
		return Medium, nil
	case "high":
		return High, nil
	case "critical":
		return Critical, nil
	default:
		return None, fmt.Errorf("invalid instrumentation level: %s", value)
	}
}

func (l Level) String() string {
	switch l {
	case None:
		return "none"
	case Low:
		return "low"
	case Medium:
		return "medium"
	case High:
		return "high"
	case Critical:
		return "critical"
	default:
		return "unknown"
	}
}
