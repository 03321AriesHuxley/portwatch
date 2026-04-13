package alert

import "context"

// Notifier is anything that can send a batch of PortEvents.
type Notifier interface {
	Send(ctx context.Context, events []PortEvent) error
}

// Dispatcher fans out PortEvents to one or more Notifiers.
type Dispatcher struct {
	notifiers []Notifier
}

// NewDispatcher creates a Dispatcher with the given notifiers.
func NewDispatcher(notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{notifiers: notifiers}
}

// Dispatch sends events to all registered notifiers.
// It continues even if one notifier fails, collecting all errors.
func (d *Dispatcher) Dispatch(ctx context.Context, events []PortEvent) []error {
	if len(events) == 0 {
		return nil
	}
	var errs []error
	for _, n := range d.notifiers {
		if err := n.Send(ctx, events); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
