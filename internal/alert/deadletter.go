package alert

import (
	"context"
	"sync"
	"time"
)

// DeadLetterNotifier wraps a Notifier and captures events that fail after all
// retries, storing them in an in-memory dead-letter queue for inspection.
type DeadLetterNotifier struct {
	inner    Notifier
	mu       sync.Mutex
	queue    []DeadLetterEntry
	maxQueue int
}

// DeadLetterEntry holds a failed batch of events and the error that caused the failure.
type DeadLetterEntry struct {
	Events    []Event
	Err       error
	FailedAt  time.Time
}

// NewDeadLetterNotifier wraps inner and captures events that inner fails to deliver.
// maxQueue is the maximum number of failed batches to retain (oldest are dropped).
func NewDeadLetterNotifier(inner Notifier, maxQueue int) *DeadLetterNotifier {
	if maxQueue <= 0 {
		maxQueue = 100
	}
	return &DeadLetterNotifier{
		inner:    inner,
		maxQueue: maxQueue,
	}
}

// Send attempts delivery via the inner notifier. On failure the events are
// captured in the dead-letter queue and the error is returned to the caller.
func (d *DeadLetterNotifier) Send(ctx context.Context, events []Event) error {
	if err := d.inner.Send(ctx, events); err != nil {
		d.mu.Lock()
		defer d.mu.Unlock()
		entry := DeadLetterEntry{
			Events:   events,
			Err:      err,
			FailedAt: time.Now(),
		}
		d.queue = append(d.queue, entry)
		if len(d.queue) > d.maxQueue {
			d.queue = d.queue[len(d.queue)-d.maxQueue:]
		}
		return err
	}
	return nil
}

// Drain returns and clears all dead-letter entries.
func (d *DeadLetterNotifier) Drain() []DeadLetterEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DeadLetterEntry, len(d.queue))
	copy(out, d.queue)
	d.queue = d.queue[:0]
	return out
}

// Len returns the current number of dead-letter entries.
func (d *DeadLetterNotifier) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.queue)
}
