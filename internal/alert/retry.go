package alert

import (
	"context"
	"fmt"
	"time"
)

// RetryNotifier wraps a Notifier and retries failed sends up to MaxAttempts times.
type RetryNotifier struct {
	inner       Notifier
	maxAttempts int
	delay       time.Duration
}

// NewRetryNotifier creates a RetryNotifier that retries up to maxAttempts times
// with the given delay between attempts.
func NewRetryNotifier(inner Notifier, maxAttempts int, delay time.Duration) *RetryNotifier {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &RetryNotifier{
		inner:       inner,
		maxAttempts: maxAttempts,
		delay:       delay,
	}
}

// Send attempts to deliver events, retrying on failure up to MaxAttempts times.
// It respects context cancellation between retry attempts.
func (r *RetryNotifier) Send(ctx context.Context, events []Event) error {
	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		lastErr = r.inner.Send(ctx, events)
		if lastErr == nil {
			return nil
		}
		if attempt < r.maxAttempts {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.delay):
			}
		}
	}
	return fmt.Errorf("all %d attempts failed: %w", r.maxAttempts, lastErr)
}
