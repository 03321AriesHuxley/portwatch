package alert

import (
	"context"
	"fmt"
)

// Enricher adds contextual key-value pairs to every Event's Meta map.
type Enricher interface {
	Enrich(e Event) map[string]string
}

// StaticEnricher returns a fixed set of key-value pairs for every event.
type StaticEnricher struct {
	Fields map[string]string
}

func (s StaticEnricher) Enrich(_ Event) map[string]string { return s.Fields }

// PortEnricher adds human-readable port context (e.g. well-known service name).
type PortEnricher struct{}

func (PortEnricher) Enrich(e Event) map[string]string {
	return map[string]string{
		"port_label": fmt.Sprintf("%s/%d", e.Entry.Protocol, e.Entry.Port),
	}
}

// enrichNotifier wraps a Notifier and applies one or more Enrichers to each
// event before forwarding.
type enrichNotifier struct {
	inner    Notifier
	enrichers []Enricher
}

// NewEnrichNotifier returns a Notifier that enriches events with additional
// metadata before delegating to inner.
func NewEnrichNotifier(inner Notifier, enrichers ...Enricher) Notifier {
	return &enrichNotifier{inner: inner, enrichers: enrichers}
}

func (n *enrichNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return n.inner.Send(ctx, events)
	}
	enriched := make([]Event, len(events))
	for i, e := range events {
		if e.Meta == nil {
			e.Meta = make(map[string]string)
		}
		for _, enr := range n.enrichers {
			for k, v := range enr.Enrich(e) {
				e.Meta[k] = v
			}
		}
		enriched[i] = e
	}
	return n.inner.Send(ctx, enriched)
}
