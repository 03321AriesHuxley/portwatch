package alert

import (
	"sync"
	"time"
)

// RateLimiter wraps a Notifier and suppresses sends that occur too frequently.
// It allows at most one notification per window duration.
type RateLimiter struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	lastSent time.Time
	now      func() time.Time
}

// NewRateLimiter returns a RateLimiter that wraps inner and enforces a minimum
// window between successive Send calls.
func NewRateLimiter(inner Notifier, window time.Duration) *RateLimiter {
	return &RateLimiter{
		inner:  inner,
		window: window,
		now:    time.Now,
	}
}

// Send forwards events to the wrapped Notifier only if the rate window has
// elapsed since the last successful send. Suppressed calls return nil.
func (r *RateLimiter) Send(events []Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(events) == 0 {
		return nil
	}

	now := r.now()
	if !r.lastSent.IsZero() && now.Sub(r.lastSent) < r.window {
		// Still within the rate-limit window; suppress this send.
		return nil
	}

	if err := r.inner.Send(events); err != nil {
		return err
	}

	r.lastSent = now
	return nil
}
