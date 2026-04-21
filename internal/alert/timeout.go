package alert

import (
	"context"
	"fmt"
	"time"
)

// TimeoutNotifier wraps a Notifier and enforces a per-Send deadline.
// If the inner notifier does not return within the configured duration,
// the context is cancelled and an error is returned.
type TimeoutNotifier struct {
	inner   Notifier
	timeout time.Duration
}

// NewTimeoutNotifier returns a Notifier that cancels the inner Send call
// if it exceeds the given timeout duration.
func NewTimeoutNotifier(inner Notifier, timeout time.Duration) *TimeoutNotifier {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &TimeoutNotifier{inner: inner, timeout: timeout}
}

// Send forwards events to the inner notifier, enforcing the configured deadline.
// It returns an error if the deadline is exceeded before the inner call returns.
func (t *TimeoutNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type result struct {
		err error
	}

	ch := make(chan result, 1)
	go func() {
		ch <- result{err: t.inner.Send(ctx, events)}
	}()

	select {
	case res := <-ch:
		return res.err
	case <-ctx.Done():
		return fmt.Errorf("timeout notifier: send exceeded %s deadline: %w", t.timeout, ctx.Err())
	}
}
