package alert

import (
	"context"
	"sync"
	"time"
)

// ReplayStore holds a fixed-size ring buffer of past events for replay.
type ReplayStore struct {
	mu     sync.Mutex
	buf    [][]Event
	max    int
	times  []time.Time
}

// NewReplayStore creates a ReplayStore that retains up to maxEntries batches.
func NewReplayStore(maxEntries int) *ReplayStore {
	if maxEntries <= 0 {
		maxEntries = 50
	}
	return &ReplayStore{max: maxEntries}
}

// Record appends a batch to the ring buffer.
func (r *ReplayStore) Record(events []Event) {
	if len(events) == 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.buf) >= r.max {
		r.buf = r.buf[1:]
		r.times = r.times[1:]
	}
	copy_ := make([]Event, len(events))
	copy(copy_, events)
	r.buf = append(r.buf, copy_)
	r.times = append(r.times, time.Now())
}

// Since returns all batches recorded after t.
func (r *ReplayStore) Since(t time.Time) [][]Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out [][]Event
	for i, ts := range r.times {
		if ts.After(t) {
			out = append(out, r.buf[i])
		}
	}
	return out
}

// NewReplayNotifier wraps a Notifier and records every sent batch.
func NewReplayNotifier(inner Notifier, store *ReplayStore) Notifier {
	return notifierFunc(func(ctx context.Context, events []Event) error {
		err := inner.Send(ctx, events)
		if err == nil {
			store.Record(events)
		}
		return err
	})
}
