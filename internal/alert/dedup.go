package alert

import (
	"sync"
	"time"
)

// Deduplicator suppresses repeated identical events within a time window.
type Deduplicator struct {
	mu     sync.Mutex
	seen   map[string]time.Time
	window time.Duration
	now    func() time.Time
}

// NewDeduplicator creates a Deduplicator that suppresses duplicate events
// within the given window duration.
func NewDeduplicator(window time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
		now:    time.Now,
	}
}

// Filter returns only events that have not been seen within the dedup window.
func (d *Deduplicator) Filter(events []Event) []Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.evict(now)

	out := make([]Event, 0, len(events))
	for _, e := range events {
		key := e.Kind + "|" + e.Entry.String()
		if _, exists := d.seen[key]; !exists {
			d.seen[key] = now
			out = append(out, e)
		}
	}
	return out
}

// evict removes entries older than the dedup window.
func (d *Deduplicator) evict(now time.Time) {
	for key, ts := range d.seen {
		if now.Sub(ts) >= d.window {
			delete(d.seen, key)
		}
	}
}
