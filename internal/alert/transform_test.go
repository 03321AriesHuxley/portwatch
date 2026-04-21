package alert

import (
	"context"
	"errors"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func makeTransformEvent(port uint16, kind string) Event {
	return Event{
		Kind: kind,
		Entry: scanner.Entry{
			Port:    port,
			Address: "127.0.0.1",
			Proto:   "tcp",
		},
	}
}

func TestTransformNotifier_EmptyEvents(t *testing.T) {
	called := false
	next := NotifierFunc(func(_ context.Context, events []Event) error {
		called = true
		return nil
	})
	n := NewTransformNotifier(func(e []Event) []Event { return e }, next)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("delegate should not be called for empty events")
	}
}

func TestTransformNotifier_TransformDropsAll(t *testing.T) {
	called := false
	next := NotifierFunc(func(_ context.Context, events []Event) error {
		called = true
		return nil
	})
	n := NewTransformNotifier(func(_ []Event) []Event { return nil }, next)
	events := []Event{makeTransformEvent(80, "added")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("delegate should not be called when transform returns empty")
	}
}

func TestTransformNotifier_ForwardsTransformed(t *testing.T) {
	var got []Event
	next := NotifierFunc(func(_ context.Context, events []Event) error {
		got = events
		return nil
	})
	transform := PortTransform(map[uint16]uint16{80: 8080})
	n := NewTransformNotifier(transform, next)
	events := []Event{makeTransformEvent(80, "added")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Entry.Port != 8080 {
		t.Fatalf("expected port 8080, got %+v", got)
	}
}

func TestPortTransform_UnmappedPassthrough(t *testing.T) {
	fn := PortTransform(map[uint16]uint16{443: 8443})
	events := []Event{makeTransformEvent(80, "added")}
	out := fn(events)
	if out[0].Entry.Port != 80 {
		t.Fatalf("expected port 80 unchanged, got %d", out[0].Entry.Port)
	}
}

func TestKindTransform_OverridesKind(t *testing.T) {
	fn := KindTransform("removed")
	events := []Event{makeTransformEvent(80, "added")}
	out := fn(events)
	if out[0].Kind != "removed" {
		t.Fatalf("expected kind 'removed', got %q", out[0].Kind)
	}
}

func TestChainTransforms_AppliesInOrder(t *testing.T) {
	portMap := PortTransform(map[uint16]uint16{80: 8080})
	kindMap := KindTransform("removed")
	chain := ChainTransforms(portMap, kindMap)
	events := []Event{makeTransformEvent(80, "added")}
	out := chain(events)
	if out[0].Entry.Port != 8080 || out[0].Kind != "removed" {
		t.Fatalf("unexpected result: %+v", out[0])
	}
}

func TestTransformNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("delegate error")
	next := NotifierFunc(func(_ context.Context, _ []Event) error {
		return sentinel
	})
	n := NewTransformNotifier(func(e []Event) []Event { return e }, next)
	events := []Event{makeTransformEvent(443, "added")}
	err := n.Send(context.Background(), events)
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
