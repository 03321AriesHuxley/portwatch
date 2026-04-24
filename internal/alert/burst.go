package alert

import (
	"context"
	"sync"
	"time"
)

// BurstNotifier allows up to MaxBurst events through immediately, then
// enforces a per-event rate (1 event per Rate duration) until the burst
// allowance refills after RefillAfter.
type BurstNotifier struct {
	inner     Notifier
	maxBurst  int
	rate      time.Duration
	refill    time.Duration
	clock     func() time.Time

	mu        sync.Mutex
	remaining int
	lastRefill time.Time
	lastSent   time.Time
}

type BurstOption func(*BurstNotifier)

func WithBurstClock(fn func() time.Time) BurstOption {
	return func(b *BurstNotifier) { b.clock = fn }
}

// NewBurstNotifier creates a notifier that allows bursts of up to maxBurst
// events, then throttles to one event per rate duration. The burst allowance
// is fully refilled after refillAfter has elapsed since the last send.
func NewBurstNotifier(inner Notifier, maxBurst int, rate, refillAfter time.Duration, opts ...BurstOption) *BurstNotifier {
	n := &BurstNotifier{
		inner:    inner,
		maxBurst: maxBurst,
		rate:     rate,
		refill:   refillAfter,
		clock:    time.Now,
	}
	for _, o := range opts {
		o(n)
	}
	n.remaining = maxBurst
	n.lastRefill = n.clock()
	return n
}

func (b *BurstNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	b.mu.Lock()
	now := b.clock()

	// Refill burst allowance if the refill window has elapsed.
	if now.Sub(b.lastSent) >= b.refill {
		b.remaining = b.maxBurst
		b.lastRefill = now
	}

	var allowed []Event
	for _, ev := range events {
		if b.remaining > 0 {
			allowed = append(allowed, ev)
			b.remaining--
		} else if b.rate > 0 && now.Sub(b.lastSent) >= b.rate {
			allowed = append(allowed, ev)
			b.lastSent = now
		}
	}

	if len(allowed) > 0 {
		b.lastSent = now
	}
	b.mu.Unlock()

	if len(allowed) == 0 {
		return nil
	}
	return b.inner.Send(ctx, allowed)
}
