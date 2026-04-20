package alert

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func makeQuotaEvents(n int) []Event {
	events := make([]Event, n)
	for i := range events {
		events[i] = Event{Kind: EventAdded, Entry: makeEntry(8080+i, "tcp", "0.0.0.0")}
	}
	return events
}

func TestQuotaNotifier_AllowsWithinQuota(t *testing.T) {
	var sent int32
	inner := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&sent, int32(len(events)))
		return nil
	})
	q := NewQuotaNotifier(inner, 5, time.Minute)

	if err := q.Send(context.Background(), makeQuotaEvents(3)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := atomic.LoadInt32(&sent); got != 3 {
		t.Fatalf("expected 3 sent, got %d", got)
	}
	if rem := q.Remaining(); rem != 2 {
		t.Fatalf("expected 2 remaining, got %d", rem)
	}
}

func TestQuotaNotifier_ExhaustsQuota(t *testing.T) {
	var sent int32
	inner := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&sent, int32(len(events)))
		return nil
	})
	q := NewQuotaNotifier(inner, 3, time.Minute)

	_ = q.Send(context.Background(), makeQuotaEvents(3))
	err := q.Send(context.Background(), makeQuotaEvents(2))
	if err == nil {
		t.Fatal("expected quota error, got nil")
	}
	if !strings.Contains(err.Error(), "quota exhausted") {
		t.Fatalf("unexpected error: %v", err)
	}
	if rem := q.Remaining(); rem != 0 {
		t.Fatalf("expected 0 remaining, got %d", rem)
	}
}

func TestQuotaNotifier_PartialBatchTruncated(t *testing.T) {
	var sent int32
	inner := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&sent, int32(len(events)))
		return nil
	})
	q := NewQuotaNotifier(inner, 4, time.Minute)
	_ = q.Send(context.Background(), makeQuotaEvents(2))
	_ = q.Send(context.Background(), makeQuotaEvents(5)) // only 2 should pass
	if got := atomic.LoadInt32(&sent); got != 4 {
		t.Fatalf("expected 4 total sent, got %d", got)
	}
}

func TestQuotaNotifier_ResetsAfterWindow(t *testing.T) {
	var sent int32
	inner := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&sent, int32(len(events)))
		return nil
	})
	now := time.Now()
	q := NewQuotaNotifier(inner, 2, 50*time.Millisecond)
	q.now = func() time.Time { return now }

	_ = q.Send(context.Background(), makeQuotaEvents(2))

	// advance past window
	q.now = func() time.Time { return now.Add(100 * time.Millisecond) }
	if err := q.Send(context.Background(), makeQuotaEvents(2)); err != nil {
		t.Fatalf("expected reset, got error: %v", err)
	}
	if got := atomic.LoadInt32(&sent); got != 4 {
		t.Fatalf("expected 4 total sent after reset, got %d", got)
	}
}

func TestQuotaNotifier_EmptyEventsSkipped(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ []Event) error {
		return errors.New("should not be called")
	})
	q := NewQuotaNotifier(inner, 5, time.Minute)
	if err := q.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error on empty send: %v", err)
	}
}
