package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func makeCheckpointEvent() Event {
	return Event{Kind: "added", Entry: makeEntry(8080, "tcp", "0.0.0.0")}
}

func TestCheckpointStore_EmptyOnStart(t *testing.T) {
	store := NewCheckpointStore()
	_, ok := store.Last("svc")
	if ok {
		t.Fatal("expected no checkpoint on fresh store")
	}
}

func TestCheckpointStore_MarkAndRetrieve(t *testing.T) {
	store := NewCheckpointStore()
	now := time.Now().Truncate(time.Second)
	store.Mark("svc", now)
	got, ok := store.Last("svc")
	if !ok {
		t.Fatal("expected checkpoint to exist after Mark")
	}
	if !got.Equal(now) {
		t.Fatalf("expected %v, got %v", now, got)
	}
}

func TestCheckpointStore_Reset(t *testing.T) {
	store := NewCheckpointStore()
	store.Mark("svc", time.Now())
	store.Reset("svc")
	_, ok := store.Last("svc")
	if ok {
		t.Fatal("expected checkpoint to be cleared after Reset")
	}
}

func TestCheckpointNotifier_MarksOnSuccess(t *testing.T) {
	store := NewCheckpointStore()
	var received []Event
	inner := NotifierFunc(func(_ context.Context, evs []Event) error {
		received = evs
		return nil
	})
	cn := NewCheckpointNotifier("svc", inner, store)
	events := []Event{makeCheckpointEvent()}
	if err := cn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected inner to receive 1 event, got %d", len(received))
	}
	_, ok := store.Last("svc")
	if !ok {
		t.Fatal("expected checkpoint to be marked after successful send")
	}
}

func TestCheckpointNotifier_DoesNotMarkOnFailure(t *testing.T) {
	store := NewCheckpointStore()
	sentinel := errors.New("send failed")
	inner := NotifierFunc(func(_ context.Context, _ []Event) error {
		return sentinel
	})
	cn := NewCheckpointNotifier("svc", inner, store)
	err := cn.Send(context.Background(), []Event{makeCheckpointEvent()})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	_, ok := store.Last("svc")
	if ok {
		t.Fatal("checkpoint must not be recorded on failure")
	}
}

func TestCheckpointNotifier_EmptyEventsSkipped(t *testing.T) {
	store := NewCheckpointStore()
	inner := NotifierFunc(func(_ context.Context, _ []Event) error { return nil })
	cn := NewCheckpointNotifier("svc", inner, store)
	if err := cn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := store.Last("svc")
	if ok {
		t.Fatal("checkpoint must not be recorded for empty event list")
	}
}
