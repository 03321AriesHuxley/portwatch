package alert

import (
	"context"
	"sync"
	"time"
)

// CooldownNotifier suppresses all alerts for a fixed duration after a burst
// threshold is reached within a sliding window.
type CooldownNotifier struct {
	inner     Notifier
	window    time.Duration
	cooldown  time.Duration
	threshold int

	mu          sync.Mutex
	events      []time.Time
	coolingUntil time.Time
}

// NewCooldownNotifier returns a notifier that enters a cooldown period once
// more than threshold events are observed within window.
func NewCooldownNotifier(inner Notifier, threshold int, window, cooldown time.Duration) *CooldownNotifier {
	return &CooldownNotifier{
		inner:     inner,
		threshold: threshold,
		window:    window,
		cooldown:  cooldown,
	}
}

func (c *CooldownNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	c.mu.Lock()
	now := time.Now()

	if now.Before(c.coolingUntil) {
		c.mu.Unlock()
		return nil
	}

	// prune old events outside the window
	cutoff := now.Add(-c.window)
	filtered := c.events[:0]
	for _, t := range c.events {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	c.events = filtered

	if len(c.events) > c.threshold {
		c.coolingUntil = now.Add(c.cooldown)
		c.events = nil
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

	return c.inner.Send(ctx, events)
}
