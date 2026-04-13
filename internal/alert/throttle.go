package alert

import (
	"context"
	"sync"
	"time"
)

// ThrottleNotifier wraps a Notifier and ensures at most one notification
// is sent per throttle window, dropping excess calls silently.
type ThrottleNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	lastSent time.Time
}

// NewThrottleNotifier creates a ThrottleNotifier that forwards to inner at
// most once per window duration.
func NewThrottleNotifier(inner Notifier, window time.Duration) *ThrottleNotifier {
	return &ThrottleNotifier{
		inner:  inner,
		window: window,
	}
}

// Send forwards events to the inner notifier only if the throttle window
// has elapsed since the last successful send.
func (t *ThrottleNotifier) Send(ctx context.Context, events []Event) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(events) == 0 {
		return nil
	}

	now := time.Now()
	if !t.lastSent.IsZero() && now.Sub(t.lastSent) < t.window {
		return nil
	}

	if err := t.inner.Send(ctx, events); err != nil {
		return err
	}

	t.lastSent = now
	return nil
}

// Reset clears the throttle state, allowing the next Send to go through
// regardless of the window.
func (t *ThrottleNotifier) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSent = time.Time{}
}
