package alert

import "context"

// TruncateNotifier limits the number of events forwarded per Send call.
// Events beyond MaxEvents are silently dropped. This is useful as a
// last-resort safety valve before a downstream sink.
type TruncateNotifier struct {
	inner     Notifier
	maxEvents int
}

// NewTruncateNotifier returns a Notifier that forwards at most maxEvents
// events to inner per call. If maxEvents <= 0 it defaults to 1.
func NewTruncateNotifier(inner Notifier, maxEvents int) *TruncateNotifier {
	if maxEvents <= 0 {
		maxEvents = 1
	}
	return &TruncateNotifier{inner: inner, maxEvents: maxEvents}
}

// Send forwards up to t.maxEvents events from events to the inner Notifier.
// If events is empty the call is skipped entirely.
func (t *TruncateNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if len(events) > t.maxEvents {
		events = events[:t.maxEvents]
	}
	return t.inner.Send(ctx, events)
}
