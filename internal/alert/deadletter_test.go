package alert

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/scanner/entry"
)

// failingNotifier always returns the provided error.
type failingNotifier struct{ err error }

func (f *failingNotifier) Send(_ context.Context, _ []Event) error { return f.err }

// succeedingNotifier always returns nil.
type succeedingNotifier struct{}

func (s *succeedingNotifier) Send(_ context.Context, _ []Event) error { return nil }

func makeDLEvent() Event {
	return Event{Kind: KindAdded, Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9090, Protocol: "tcp"}}
}

func TestDeadLetterNotifier_SuccessDoesNotQueue(t *testing.T) {
	dl := NewDeadLetterNotifier(&succeedingNotifier{}, 10)
	events := []Event{makeDLEvent()}

	if err := dl.Send(context.Background(), events); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if dl.Len() != 0 {
		t.Errorf("expected empty dead-letter queue, got %d entries", dl.Len())
	}
}

func TestDeadLetterNotifier_CapturesFailure(t *testing.T) {
	sentinel := errors.New("send failed")
	dl := NewDeadLetterNotifier(&failingNotifier{err: sentinel}, 10)
	events := []Event{makeDLEvent()}

	err := dl.Send(context.Background(), events)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if dl.Len() != 1 {
		t.Fatalf("expected 1 dead-letter entry, got %d", dl.Len())
	}
}

func TestDeadLetterNotifier_DrainClearsQueue(t *testing.T) {
	dl := NewDeadLetterNotifier(&failingNotifier{err: errors.New("boom")}, 10)
	events := []Event{makeDLEvent()}

	_ = dl.Send(context.Background(), events)
	_ = dl.Send(context.Background(), events)

	entries := dl.Drain()
	if len(entries) != 2 {
		t.Fatalf("expected 2 drained entries, got %d", len(entries))
	}
	if dl.Len() != 0 {
		t.Errorf("expected empty queue after drain, got %d", dl.Len())
	}
	for _, e := range entries {
		if e.FailedAt.IsZero() {
			t.Error("expected FailedAt to be set")
		}
		if e.Err == nil {
			t.Error("expected Err to be set")
		}
	}
}

func TestDeadLetterNotifier_RespectsMaxQueue(t *testing.T) {
	dl := NewDeadLetterNotifier(&failingNotifier{err: errors.New("x")}, 3)
	events := []Event{makeDLEvent()}

	for i := 0; i < 5; i++ {
		_ = dl.Send(context.Background(), events)
		time.Sleep(time.Millisecond)
	}

	if dl.Len() != 3 {
		t.Errorf("expected max 3 entries, got %d", dl.Len())
	}
}

func TestDeadLetterNotifier_DefaultMaxQueue(t *testing.T) {
	dl := NewDeadLetterNotifier(&failingNotifier{err: errors.New("x")}, 0)
	if dl.maxQueue != 100 {
		t.Errorf("expected default maxQueue=100, got %d", dl.maxQueue)
	}
}
