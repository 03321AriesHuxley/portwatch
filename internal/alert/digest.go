package alert

import (
	"context"
	"sync"
	"time"
)

// DigestNotifier accumulates events over a fixed window and forwards a single
// digest batch to the inner notifier at the end of each window. Unlike
// SummaryNotifier (which formats a human-readable summary), DigestNotifier
// forwards the raw accumulated events so downstream notifiers can process them.
type DigestNotifier struct {
	inner   Notifier
	window  time.Duration
	clock   func() time.Time
	mu      sync.Mutex
	buf     []Event
	windowStart time.Time
}

// NewDigestNotifier creates a DigestNotifier that collects events for the given
// window duration before forwarding them all at once to inner.
func NewDigestNotifier(inner Notifier, window time.Duration) *DigestNotifier {
	return &DigestNotifier{
		inner:  inner,
		window: window,
		clock:  time.Now,
	}
}

// Send buffers events. When the current wall-clock window expires the entire
// accumulated buffer is flushed to the inner notifier and the window resets.
func (d *DigestNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	d.mu.Lock()
	now := d.clock()
	if d.windowStart.IsZero() {
		d.windowStart = now
	}

	d.buf = append(d.buf, events...)

	if now.Sub(d.windowStart) < d.window {
		d.mu.Unlock()
		return nil
	}

	// Window has elapsed — flush.
	batch := d.buf
	d.buf = nil
	d.windowStart = now
	d.mu.Unlock()

	return d.inner.Send(ctx, batch)
}

// Flush immediately forwards any buffered events to the inner notifier,
// regardless of whether the window has elapsed.
func (d *DigestNotifier) Flush(ctx context.Context) error {
	d.mu.Lock()
	batch := d.buf
	d.buf = nil
	d.windowStart = time.Time{}
	d.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}
	return d.inner.Send(ctx, batch)
}
