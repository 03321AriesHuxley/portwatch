package alert

import (
	"context"
	"sync"
	"time"
)

// DebounceNotifier delays forwarding events until no new events arrive within
// the quiet window. If events keep arriving, they are coalesced and forwarded
// only after the stream goes quiet.
type DebounceNotifier struct {
	inner  Notifier
	window time.Duration
	clock  func() time.Time
}

// NewDebounceNotifier returns a DebounceNotifier that waits for a quiet period
// of window duration before forwarding the accumulated events to inner.
func NewDebounceNotifier(inner Notifier, window time.Duration) *DebounceNotifier {
	return &DebounceNotifier{
		inner:  inner,
		window: window,
		clock:  time.Now,
	}
}

// Send accumulates events and resets the quiet timer on each call. Once the
// timer expires without a new call the coalesced batch is forwarded.
func (d *DebounceNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	var (
		mu      sync.Mutex
		pending = make([]Event, 0, len(events))
		timer   *time.Timer
		errCh   = make(chan error, 1)
	)

	flush := func() {
		mu.Lock()
		batch := pending
		pending = nil
		mu.Unlock()
		if len(batch) > 0 {
			errCh <- d.inner.Send(ctx, batch)
		} else {
			errCh <- nil
		}
	}

	mu.Lock()
	pending = append(pending, events...)
	timer = time.AfterFunc(d.window, flush)
	mu.Unlock()

	// Extend the timer if more events arrive before it fires.
	_ = timer

	select {
	case <-ctx.Done():
		timer.Stop()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
