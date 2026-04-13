package alert

import "context"

// Notifier is the interface implemented by all alerting backends.
// Send delivers a batch of port-change events to the backend.
// Implementations must honour context cancellation.
type Notifier interface {
	Send(ctx context.Context, events []Event) error
}

// MultiNotifier fans out a Send call to multiple Notifier implementations.
// It collects all errors and returns the first non-nil one, but always
// attempts every notifier regardless of prior failures.
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier creates a MultiNotifier that delegates to each provided Notifier.
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Send calls every contained Notifier and returns the first error encountered.
func (m *MultiNotifier) Send(ctx context.Context, events []Event) error {
	var firstErr error
	for _, n := range m.notifiers {
		if err := n.Send(ctx, events); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
