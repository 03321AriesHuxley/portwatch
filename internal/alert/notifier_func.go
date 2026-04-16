package alert

import "context"

// notifierFunc is a function adapter for the Notifier interface.
type notifierFunc func(ctx context.Context, events []Event) error

func (f notifierFunc) Send(ctx context.Context, events []Event) error {
	return f(ctx, events)
}
