package alert

import (
	"context"
	"testing"
	"time"
)

type recordingNotifier struct {
	calls [][]Event
}

func (r *recordingNotifier) Send(_ context.Context, events []Event) error {
	r.calls = append(r.calls, events)
	return nil
}

func makeSummaryEvents(n int) []Event {
	events := make([]Event, n)
	for i := range events {
		events[i] = Event{Kind: KindAdded}
	}
	return events
}

func TestSummaryNotifier_BuffersWithinWindow(t *testing.T) {
	rec := &recordingNotifier{}
	sn := NewSummaryNotifier(rec, 10*time.Second)

	_ = sn.Send(context.Background(), makeSummaryEvents(3))

	if len(rec.calls) != 0 {
		t.Fatalf("expected 0 forwarded calls within window, got %d", len(rec.calls))
	}
	if len(sn.buf) != 3 {
		t.Fatalf("expected 3 buffered events, got %d", len(sn.buf))
	}
}

func TestSummaryNotifier_ForwardsAfterWindow(t *testing.T) {
	rec := &recordingNotifier{}
	sn := NewSummaryNotifier(rec, 0) // zero window: always flush

	_ = sn.Send(context.Background(), makeSummaryEvents(4))

	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 forwarded call, got %d", len(rec.calls))
	}
	if len(rec.calls[0]) != 4 {
		t.Fatalf("expected 4 events in call, got %d", len(rec.calls[0]))
	}
	if len(sn.buf) != 0 {
		t.Fatalf("expected empty buffer after flush, got %d", len(sn.buf))
	}
}

func TestSummaryNotifier_FlushDrainsBuffer(t *testing.T) {
	rec := &recordingNotifier{}
	sn := NewSummaryNotifier(rec, 10*time.Second)

	_ = sn.Send(context.Background(), makeSummaryEvents(2))
	_ = sn.Flush(context.Background())

	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 forwarded call after explicit flush, got %d", len(rec.calls))
	}
	if len(sn.buf) != 0 {
		t.Fatalf("expected empty buffer, got %d", len(sn.buf))
	}
}

func TestSummaryNotifier_EmptyEventsNoOp(t *testing.T) {
	rec := &recordingNotifier{}
	sn := NewSummaryNotifier(rec, 0)

	_ = sn.Send(context.Background(), []Event{})

	if len(rec.calls) != 0 {
		t.Fatalf("expected no calls for empty events, got %d", len(rec.calls))
	}
}

func TestSummaryNotifier_FlushEmptyBufferNoOp(t *testing.T) {
	rec := &recordingNotifier{}
	sn := NewSummaryNotifier(rec, 10*time.Second)

	_ = sn.Flush(context.Background())

	if len(rec.calls) != 0 {
		t.Fatalf("expected no calls flushing empty buffer, got %d", len(rec.calls))
	}
}
