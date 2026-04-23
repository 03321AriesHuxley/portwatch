package alert

import (
	"context"
	"sync"
	"time"
)

// CacheNotifier wraps a Notifier and caches the last successful send.
// If the inner notifier fails, the cached events are preserved and
// the caller can inspect them via Cached().
type CacheNotifier struct {
	inner    Notifier
	mu       sync.RWMutex
	cached   []Event
	cachedAt time.Time
	ttl      time.Duration
	clock    func() time.Time
}

// NewCacheNotifier wraps inner, caching the last successful batch for ttl.
func NewCacheNotifier(inner Notifier, ttl time.Duration) *CacheNotifier {
	return &CacheNotifier{
		inner: inner,
		ttl:   ttl,
		clock: time.Now,
	}
}

// Send forwards events to the inner notifier. On success the events are
// stored in the cache, replacing any previous entry. On failure the
// existing cache is left untouched and the error is returned.
func (c *CacheNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if err := c.inner.Send(ctx, events); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]Event, len(events))
	copy(cp, events)
	c.cached = cp
	c.cachedAt = c.clock()
	return nil
}

// Cached returns the most recently cached events and the time they were
// stored. If the cache is empty or has expired the returned slice is nil.
func (c *CacheNotifier) Cached() ([]Event, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cached == nil {
		return nil, time.Time{}
	}
	if c.ttl > 0 && c.clock().Sub(c.cachedAt) > c.ttl {
		return nil, time.Time{}
	}
	cp := make([]Event, len(c.cached))
	copy(cp, c.cached)
	return cp, c.cachedAt
}

// Invalidate clears the cache immediately.
func (c *CacheNotifier) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cached = nil
	c.cachedAt = time.Time{}
}
