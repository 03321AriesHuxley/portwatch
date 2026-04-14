package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeSnapEvent() alert.Event {
	return alert.Event{
		Kind:  alert.EventAdded,
		Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9000, Protocol: "tcp"},
	}
}

type stubSnapNotifier struct{ err error }

func (s *stubSnapNotifier) Send(_ context.Context, events []alert.Event) error {
	return s.err
}

func TestSnapshotNotifier_RecordsSuccess(t *testing.T) {
	store := alert.NewSnapshotStore()
	inner := &stubSnapNotifier{}
	sn := alert.NewSnapshotNotifier("test", inner, store)

	_ = sn.Send(context.Background(), []alert.Event{makeSnapEvent(), makeSnapEvent()})

	snap, ok := store.Get("test")
	if !ok {
		t.Fatal("expected snapshot to be recorded")
	}
	if snap.Sent != 2 {
		t.Errorf("expected Sent=2, got %d", snap.Sent)
	}
	if snap.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", snap.Failed)
	}
	if snap.LastSent.IsZero() {
		t.Error("expected LastSent to be set")
	}
}

func TestSnapshotNotifier_RecordsFailure(t *testing.T) {
	store := alert.NewSnapshotStore()
	inner := &stubSnapNotifier{err: errors.New("send failed")}
	sn := alert.NewSnapshotNotifier("test", inner, store)

	_ = sn.Send(context.Background(), []alert.Event{makeSnapEvent()})

	snap, ok := store.Get("test")
	if !ok {
		t.Fatal("expected snapshot to be recorded")
	}
	if snap.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", snap.Failed)
	}
	if snap.LastError == nil {
		t.Error("expected LastError to be set")
	}
}

func TestSnapshotNotifier_RecordsFiltered(t *testing.T) {
	store := alert.NewSnapshotStore()
	inner := &stubSnapNotifier{}
	sn := alert.NewSnapshotNotifier("test", inner, store)

	_ = sn.Send(context.Background(), nil)

	snap, ok := store.Get("test")
	if !ok {
		t.Fatal("expected snapshot to be recorded")
	}
	if snap.Filtered != 1 {
		t.Errorf("expected Filtered=1, got %d", snap.Filtered)
	}
}

func TestSnapshotStore_All(t *testing.T) {
	store := alert.NewSnapshotStore()
	store.Record(alert.NotifierSnapshot{Name: "a", Sent: 1})
	store.Record(alert.NotifierSnapshot{Name: "b", Sent: 2, LastSent: time.Now()})

	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(all))
	}
}

func TestSnapshotStore_GetMissing(t *testing.T) {
	store := alert.NewSnapshotStore()
	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("expected false for missing key")
	}
}
