package alert

import (
	"errors"
	"testing"
	"time"
)

// stubNotifier records every Send call made to it.
type stubNotifier struct {
	calls  int
	events [][]Event
	err    error
}

func (s *stubNotifier) Send(events []Event) error {
	if s.err != nil {
		return s.err
	}
	s.calls++
	s.events = append(s.events, events)
	return nil
}

func sampleEventsRL() []Event {
	return []Event{{Kind: "added", Port: 8080, Proto: "tcp"}}
}

func TestRateLimiter_AllowsFirstSend(t *testing.T) {
	stub := &stubNotifier{}
	rl := NewRateLimiter(stub, 10*time.Second)

	if err := rl.Send(sampleEventsRL()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRateLimiter_SuppressesWithinWindow(t *testing.T) {
	stub := &stubNotifier{}
	rl := NewRateLimiter(stub, 10*time.Second)

	base := time.Now()
	rl.now = func() time.Time { return base }

	_ = rl.Send(sampleEventsRL())

	// Advance only 5 seconds — still inside the window.
	rl.now = func() time.Time { return base.Add(5 * time.Second) }
	_ = rl.Send(sampleEventsRL())

	if stub.calls != 1 {
		t.Fatalf("expected 1 call (suppressed second), got %d", stub.calls)
	}
}

func TestRateLimiter_AllowsAfterWindow(t *testing.T) {
	stub := &stubNotifier{}
	rl := NewRateLimiter(stub, 10*time.Second)

	base := time.Now()
	rl.now = func() time.Time { return base }
	_ = rl.Send(sampleEventsRL())

	// Advance beyond the window.
	rl.now = func() time.Time { return base.Add(11 * time.Second) }
	_ = rl.Send(sampleEventsRL())

	if stub.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", stub.calls)
	}
}

func TestRateLimiter_EmptyEventsSkipped(t *testing.T) {
	stub := &stubNotifier{}
	rl := NewRateLimiter(stub, 10*time.Second)

	if err := rl.Send([]Event{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls != 0 {
		t.Fatalf("expected 0 calls for empty events, got %d", stub.calls)
	}
}

func TestRateLimiter_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("send failed")
	stub := &stubNotifier{err: sentinel}
	rl := NewRateLimiter(stub, 10*time.Second)

	err := rl.Send(sampleEventsRL())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	// lastSent must NOT be updated on error.
	if !rl.lastSent.IsZero() {
		t.Fatal("lastSent should remain zero after a failed send")
	}
}
