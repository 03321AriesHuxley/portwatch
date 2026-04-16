package alert

import "context"

// Priority represents the severity level of an alert.
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

// PriorityNotifier wraps a set of notifiers keyed by priority level.
// Events at or above the threshold are forwarded to the inner notifier.
type PriorityNotifier struct {
	threshold Priority
	inner     Notifier
	priority  func(Event) Priority
}

// PriorityOption configures a PriorityNotifier.
type PriorityOption func(*PriorityNotifier)

// WithPriorityFunc sets a custom function to determine event priority.
func WithPriorityFunc(fn func(Event) Priority) PriorityOption {
	return func(p *PriorityNotifier) {
		p.priority = fn
	}
}

// NewPriorityNotifier creates a notifier that only forwards events meeting
// the minimum priority threshold.
func NewPriorityNotifier(threshold Priority, inner Notifier, opts ...PriorityOption) *PriorityNotifier {
	pn := &PriorityNotifier{
		threshold: threshold,
		inner:     inner,
		priority:  defaultPriority,
	}
	for _, o := range opts {
		o(pn)
	}
	return pn
}

// Send filters events by priority and forwards qualifying ones.
func (pn *PriorityNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	var filtered []Event
	for _, e := range events {
		if pn.priority(e) >= pn.threshold {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return pn.inner.Send(ctx, filtered)
}

// defaultPriority assigns priority based on event kind:
// added ports are Medium, removed ports are High.
func defaultPriority(e Event) Priority {
	if e.Kind == KindRemoved {
		return PriorityHigh
	}
	return PriorityMedium
}
