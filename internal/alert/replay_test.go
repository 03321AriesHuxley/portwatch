package alert

import (
	"context"
	"errors"
	"testing"
	"time"
)

func makeReplayEvents(port uint16) []Event {
	return []Event{{Kind: EventAdded, Entry: makeEntry(port)}}
}

func TestReplayStore_EmptyOnStart(t *testing.T) {
	s := NewReplayStore(10)
	got := s.Since(time.Now().Add(-time.Hour))
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestReplayStore_RecordsAndRetrieves(t *testing.T) {
	s := NewReplayStore(10)
	before := time.Now()
	s.Record(makeReplayEvents(8080))
	got := s.Since(before)
	if len(got) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(got))
	}
}

func TestReplayStore_SinceFiltersOld(t *testing.T) {
	s := NewReplayStore(10)
	s.Record(makeReplayEvents(8080))
	future := time.Now().Add(time.Hour)
	got := s.Since(future)
	if len(got) != 0 {
		t.Fatalf("expected 0, got %d", len(got))
	}
}

func TestReplayStore_RingBufferEvicts(t *testing.T) {
	s := NewReplayStore(2)
	s.Record(makeReplayEvents(1))
	s.Record(makeReplayEvents(2))
	s.Record(makeReplayEvents(3))
	got := s.Since(time.Now().Add(-time.Hour))
	if len(got) != 2 {
		t.Fatalf("expected 2 after eviction, got %d", len(got))
	}
}

func TestReplayNotifier_RecordsOnSuccess(t *testing.T) {
	store := NewReplayStore(10)
	inner := notifierFunc(func(_ context.Context, _ []Event) error { return nil })
	n := NewReplayNotifier(inner, store)
	before := time.Now()
	_ = n.Send(context.Background(), makeReplayEvents(9090))
	got := store.Since(before)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
}

func TestReplayNotifier_SkipsOnError(t *testing.T) {
	store := NewReplayStore(10)
	inner := notifierFunc(func(_ context.Context, _ []Event) error { return errors.New("fail") })
	n := NewReplayNotifier(inner, store)
	before := time.Now()
	_ = n.Send(context.Background(), makeReplayEvents(9090))
	got := store.Since(before)
	if len(got) != 0 {
		t.Fatalf("expected 0 on error, got %d", len(got))
	}
}
