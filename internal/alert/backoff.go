package alert

import (
	"context"
	"math"
	"time"
)

// BackoffNotifier wraps a Notifier and applies exponential backoff between
// retry attempts, up to a configurable maximum delay.
type BackoffNotifier struct {
	inner      Notifier
	initialDelay time.Duration
	maxDelay   time.Duration
	multiplier float64
}

// BackoffOption configures a BackoffNotifier.
type BackoffOption func(*BackoffNotifier)

// WithInitialDelay sets the starting delay for backoff.
func WithInitialDelay(d time.Duration) BackoffOption {
	return func(b *BackoffNotifier) { b.initialDelay = d }
}

// WithMaxDelay sets the ceiling for backoff delay.
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(b *BackoffNotifier) { b.maxDelay = d }
}

// WithMultiplier sets the exponential growth factor (default 2.0).
func WithMultiplier(m float64) BackoffOption {
	return func(b *BackoffNotifier) { b.multiplier = m }
}

// NewBackoffNotifier returns a Notifier that retries with exponential backoff.
func NewBackoffNotifier(inner Notifier, maxAttempts int, opts ...BackoffOption) Notifier {
	b := &BackoffNotifier{
		inner:        inner,
		initialDelay: 100 * time.Millisecond,
		maxDelay:     30 * time.Second,
		multiplier:   2.0,
	}
	for _, o := range opts {
		o(b)
	}
	return notifierFunc(func(ctx context.Context, events []Event) error {
		var err error
		delay := b.initialDelay
		for attempt := 0; attempt < maxAttempts; attempt++ {
			err = b.inner.Send(ctx, events)
			if err == nil {
				return nil
			}
			if attempt == maxAttempts-1 {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			next := time.Duration(math.Min(
				float64(delay)*b.multiplier,
				float64(b.maxDelay),
			))
			delay = next
		}
		return err
	})
}

// notifierFunc adapts a function to the Notifier interface.
type notifierFunc func(ctx context.Context, events []Event) error

func (f notifierFunc) Send(ctx context.Context, events []Event) error {
	return f(ctx, events)
}
