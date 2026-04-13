package daemon

import (
	"context"
	"time"
)

// Ticker wraps a time.Ticker and provides a channel-based
// mechanism for driving periodic scans within the daemon.
type Ticker struct {
	ticker   *time.Ticker
	interval time.Duration
}

// NewTicker creates a Ticker that fires at the given interval.
func NewTicker(interval time.Duration) *Ticker {
	return &Ticker{
		ticker:   time.NewTicker(interval),
		interval: interval,
	}
}

// C returns the underlying tick channel.
func (t *Ticker) C() <-chan time.Time {
	return t.ticker.C
}

// Stop halts the ticker, releasing associated resources.
func (t *Ticker) Stop() {
	t.ticker.Stop()
}

// RunEvery calls fn on every tick until ctx is cancelled.
// It returns ctx.Err() when the context is done.
func RunEvery(ctx context.Context, interval time.Duration, fn func()) error {
	t := NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C():
			fn()
		}
	}
}
