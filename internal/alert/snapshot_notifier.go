package alert

import (
	"context"
	"time"
)

// SnapshotNotifier wraps a Notifier and records a health snapshot after every Send.
type SnapshotNotifier struct {
	name  string
	inner Notifier
	store *SnapshotStore
	sent  int64
	failed int64
	filtered int64
}

// NewSnapshotNotifier wraps inner, recording stats under name into store.
func NewSnapshotNotifier(name string, inner Notifier, store *SnapshotStore) *SnapshotNotifier {
	return &SnapshotNotifier{
		name:  name,
		inner: inner,
		store: store,
	}
}

// Send forwards events to the inner notifier and records the outcome.
func (sn *SnapshotNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		sn.filtered++
		sn.store.Record(NotifierSnapshot{
			Name:     sn.name,
			Sent:     sn.sent,
			Failed:   sn.failed,
			Filtered: sn.filtered,
		})
		return nil
	}

	err := sn.inner.Send(ctx, events)
	if err != nil {
		sn.failed++
		sn.store.Record(NotifierSnapshot{
			Name:      sn.name,
			Sent:      sn.sent,
			Failed:    sn.failed,
			Filtered:  sn.filtered,
			LastError: err,
		})
		return err
	}

	sn.sent += int64(len(events))
	sn.store.Record(NotifierSnapshot{
		Name:     sn.name,
		Sent:     sn.sent,
		Failed:   sn.failed,
		Filtered: sn.filtered,
		LastSent: time.Now(),
	})
	return nil
}
