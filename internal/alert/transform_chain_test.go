package alert

import (
	"context"
	"testing"
)

// TestTransform_WithFilter verifies that a transform stage followed by a
// filter stage correctly narrows the event set.
func TestTransform_WithFilter(t *testing.T) {
	var received []Event
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		received = append(received, events...)
		return nil
	})

	// Filter passes only port 9090.
	filtered := NewFilter([]uint16{80}, nil, sink)
	// Transform rewrites port 80 → 9090 so it survives the filter.
	transformed := NewTransformNotifier(PortTransform(map[uint16]uint16{80: 9090}), filtered)

	events := []Event{
		makeTransformEvent(80, "added"),
		makeTransformEvent(443, "added"),
	}
	if err := transformed.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Port 80 was mapped to 9090 (not excluded), port 443 is excluded.
	if len(received) != 1 || received[0].Entry.Port != 9090 {
		t.Fatalf("expected 1 event with port 9090, got %+v", received)
	}
}

// TestTransform_ChainedTransforms validates that multiple transform notifiers
// compose correctly when nested.
func TestTransform_ChainedTransforms(t *testing.T) {
	var received []Event
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		received = append(received, events...)
		return nil
	})

	// Inner: override kind to "removed".
	inner := NewTransformNotifier(KindTransform("removed"), sink)
	// Outer: remap port 22 → 2222.
	outer := NewTransformNotifier(PortTransform(map[uint16]uint16{22: 2222}), inner)

	events := []Event{makeTransformEvent(22, "added")}
	if err := outer.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].Entry.Port != 2222 {
		t.Fatalf("expected port 2222, got %d", received[0].Entry.Port)
	}
	if received[0].Kind != "removed" {
		t.Fatalf("expected kind 'removed', got %q", received[0].Kind)
	}
}
