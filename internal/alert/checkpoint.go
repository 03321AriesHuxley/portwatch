package alert

import (
	"context"
	"sync"
	"time"
)

// CheckpointStore records the last successful send time per notifier name,
// allowing downstream components to resume or skip stale work after restart.
type CheckpointStore struct {
	mu      sync.RWMutex
	records map[string]time.Time
}

// NewCheckpointStore returns an initialised CheckpointStore.
func NewCheckpointStore() *CheckpointStore {
	return &CheckpointStore{
		records: make(map[string]time.Time),
	}
}

// Mark records a successful send at the given time for name.
func (c *CheckpointStore) Mark(name string, at time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records[name] = at
}

// Last returns the last recorded time for name, and whether a record exists.
func (c *CheckpointStore) Last(name string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.records[name]
	return t, ok
}

// Reset removes the checkpoint record for name.
func (c *CheckpointStore) Reset(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.records, name)
}

// CheckpointNotifier wraps a Notifier and marks a checkpoint on every
// successful Send call.
type CheckpointNotifier struct {
	name  string
	inner Notifier
	store *CheckpointStore
	now   func() time.Time
}

// NewCheckpointNotifier returns a CheckpointNotifier that records successful
// sends in store under name.
func NewCheckpointNotifier(name string, inner Notifier, store *CheckpointStore) *CheckpointNotifier {
	return &CheckpointNotifier{
		name:  name,
		inner: inner,
		store: store,
		now:   time.Now,
	}
}

// Send delegates to the inner Notifier and, on success, marks a checkpoint.
func (c *CheckpointNotifier) Send(ctx context.Context, events []Event) error {
	if err := c.inner.Send(ctx, events); err != nil {
		return err
	}
	if len(events) > 0 {
		c.store.Mark(c.name, c.now())
	}
	return nil
}
