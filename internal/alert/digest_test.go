package alert

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func makeDigestEvent(port uint16, kind string) Event {
	return Event{Kind: kind, Entry: testEntry(port)}
}

func testEntry(port uint16) interface{} {
	return port // minimal stand-in; real code uses scanner.Entry
}

// captureNotifier records every batch it receives.
type captureNotifier struct {
	mu     sync.Mutex
	batches [][]Event
	err    error
}

func (c *captureNotifier) Send(_ context.Context, events []Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batches = append(c.batches, events)
	return c.err
}

func TestDigestNotifier_BuffersWithinWindow(t *testing.T) {
	cap := &captureNotifier{}
	now := time.Now()
	d := NewDigestNotifier(cap, 5*time.Second)
	d.clock = func() time.Time { return now }

	_ = d.Send(context.Background(), []Event{makeDigestEvent(80, "added")})
	_ = d.Send(context.Background(), []Event{makeDigestEvent(443, "added")})

	cap.mu.Lock()
	got := len(cap.batches)
	cap.mu.Unlock()

	if got != 0 {
		t.Fatalf("expected 0 batches within window, got %d", got)
	}
}

func TestDigestNotifier_FlushesAfterWindow(t *testing.T) {
	cap := &captureNotifier{}
	now := time.Now()
	d := NewDigestNotifier(cap, 1*time.Second)
	d.clock = func() time.Time { return now }

	_ = d.Send(context.Background(), []Event{makeDigestEvent(80, "added")})

	// Advance clock past window.
	now = now.Add(2 * time.Second)
	_ = d.Send(context.Background(), []Event{makeDigestEvent(443, "added")})

	cap.mu.Lock()
	got := len(cap.batches)
	cap.mu.Unlock()

	if got != 1 {
		t.Fatalf("expected 1 flushed batch, got %d", got)
	}
	if len(cap.batches[0]) != 2 {
		t.Fatalf("expected 2 events in batch, got %d", len(cap.batches[0]))
	}
}

func TestDigestNotifier_ManualFlush(t *testing.T) {
	cap := &captureNotifier{}
	d := NewDigestNotifier(cap, 1*time.Hour)

	_ = d.Send(context.Background(), []Event{makeDigestEvent(22, "added")})
	_ = d.Flush(context.Background())

	cap.mu.Lock()
	got := len(cap.batches)
	cap.mu.Unlock()

	if got != 1 {
		t.Fatalf("expected 1 batch after manual flush, got %d", got)
	}
}

func TestDigestNotifier_EmptyEventsSkipped(t *testing.T) {
	cap := &captureNotifier{}
	d := NewDigestNotifier(cap, 1*time.Second)

	if err := d.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cap.mu.Lock()
	got := len(cap.batches)
	cap.mu.Unlock()

	if got != 0 {
		t.Fatalf("expected no batches for empty send, got %d", got)
	}
}

func TestDigestNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	cap := &captureNotifier{err: sentinel}
	now := time.Now()
	d := NewDigestNotifier(cap, 1*time.Second)
	d.clock = func() time.Time { return now }

	_ = d.Send(context.Background(), []Event{makeDigestEvent(80, "added")})
	now = now.Add(2 * time.Second)
	err := d.Send(context.Background(), []Event{makeDigestEvent(443, "added")})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	_ = cmp.Diff // ensure import used
}
