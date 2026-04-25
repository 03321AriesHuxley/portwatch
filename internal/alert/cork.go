package alert

import (
	"context"
	"sync"
)

// CorkNotifier buffers events while corked and flushes them when uncorked.
// This is useful for holding alerts during maintenance windows or startup.
type CorkNotifier struct {
	mu     sync.Mutex
	corked bool
	buf    []Event
	next   Notifier
}

// NewCorkNotifier creates a CorkNotifier wrapping next.
// If startCorked is true, events are buffered until Uncork is called.
func NewCorkNotifier(next Notifier, startCorked bool) *CorkNotifier {
	return &CorkNotifier{
		next:   next,
		corked: startCorked,
	}
}

// Cork enables buffering; subsequent Send calls will buffer events.
func (c *CorkNotifier) Cork() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.corked = true
}

// Uncork disables buffering and flushes any buffered events to the inner
// notifier. The provided context is used for the flush Send call.
func (c *CorkNotifier) Uncork(ctx context.Context) error {
	c.mu.Lock()
	c.corked = false
	flushed := c.buf
	c.buf = nil
	c.mu.Unlock()

	if len(flushed) == 0 {
		return nil
	}
	return c.next.Send(ctx, flushed)
}

// Drain discards all buffered events without forwarding them.
func (c *CorkNotifier) Drain() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf = nil
}

// Send buffers events if corked, otherwise delegates immediately.
func (c *CorkNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	c.mu.Lock()
	if c.corked {
		c.buf = append(c.buf, events...)
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()
	return c.next.Send(ctx, events)
}

// Buffered returns a copy of the currently buffered events.
func (c *CorkNotifier) Buffered() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Event, len(c.buf))
	copy(out, c.buf)
	return out
}
