package dnsres

import (
	"sync"
	"time"
)

// EventType identifies the kind of resolver event.
type EventType string

const (
	EventCycleStart     EventType = "cycle_start"
	EventCycleComplete  EventType = "cycle_complete"
	EventResolveSuccess EventType = "resolve_success"
	EventResolveFailure EventType = "resolve_failure"
	EventInconsistent   EventType = "inconsistent"
)

// ResolverEvent captures resolver activity for observers.
type ResolverEvent struct {
	Type          EventType
	Time          time.Time
	Hostname      string
	Server        string
	Duration      time.Duration
	Error         string
	Addresses     []string
	Consistent    *bool
	HostnameCount int
	ServerCount   int
	Source        string
}

type eventBus struct {
	mu   sync.RWMutex
	subs map[chan ResolverEvent]struct{}
}

func newEventBus() *eventBus {
	return &eventBus{
		subs: make(map[chan ResolverEvent]struct{}),
	}
}

func (b *eventBus) subscribe(buffer int) (<-chan ResolverEvent, func()) {
	if buffer <= 0 {
		buffer = 64
	}

	ch := make(chan ResolverEvent, buffer)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	return ch, func() {
		b.mu.Lock()
		delete(b.subs, ch)
		b.mu.Unlock()
		close(ch)
	}
}

func (b *eventBus) publish(event ResolverEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subs {
		select {
		case ch <- event:
		default:
		}
	}
}
