package alert

import (
	"context"
	"testing"

	"github.com/netwatch/portwatch/internal/scanner"
)

func makePriorityEvent(kind Kind, port uint16) Event {
	return Event{
		Kind: kind,
		Entry: scanner.Entry{Port: port, Address: "0.0.0.0", Protocol: "tcp"},
	}
}

type captureNotifier struct {
	got []Event
}

func (c *captureNotifier) Send(_ context.Context, events []Event) error {
	c.got = append(c.got, events...)
	return nil
}

func TestPriorityNotifier_ForwardsAboveThreshold(t *testing.T) {
	cap := &captureNotifier{}
	pn := NewPriorityNotifier(PriorityHigh, cap)

	events := []Event{
		makePriorityEvent(KindRemoved, 443), // High
		makePriorityEvent(KindAdded, 8080),  // Medium
	}
	if err := pn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(cap.got))
	}
	if cap.got[0].Kind != KindRemoved {
		t.Errorf("expected KindRemoved, got %v", cap.got[0].Kind)
	}
}

func TestPriorityNotifier_ForwardsAllAtLowThreshold(t *testing.T) {
	cap := &captureNotifier{}
	pn := NewPriorityNotifier(PriorityLow, cap)

	events := []Event{
		makePriorityEvent(KindAdded, 80),
		makePriorityEvent(KindRemoved, 22),
	}
	if err := pn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.got) != 2 {
		t.Errorf("expected 2 events, got %d", len(cap.got))
	}
}

func TestPriorityNotifier_EmptyEvents(t *testing.T) {
	cap := &captureNotifier{}
	pn := NewPriorityNotifier(PriorityMedium, cap)
	if err := pn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.got) != 0 {
		t.Errorf("expected no events forwarded")
	}
}

func TestPriorityNotifier_CustomPriorityFunc(t *testing.T) {
	cap := &captureNotifier{}
	// Always assign Low priority
	pn := NewPriorityNotifier(PriorityHigh, cap, WithPriorityFunc(func(Event) Priority {
		return PriorityLow
	}))

	events := []Event{makePriorityEvent(KindRemoved, 9000)}
	if err := pn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.got) != 0 {
		t.Errorf("expected event to be filtered by custom priority func")
	}
}
