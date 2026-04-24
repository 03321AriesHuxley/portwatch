package alert_test

import (
	"context"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeNormalizeEvent(addr, proto string) alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{
			Addr:  addr,
			Port:  8080,
			Proto: proto,
		},
	}
}

func TestNormalizeNotifier_EmptyEvents(t *testing.T) {
	called := false
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	n := alert.NewNormalizeNotifier(inner, nil)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("inner should not be called for empty events")
	}
}

func TestNormalizeNotifier_LowercasesAddress(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewNormalizeNotifier(inner, nil)
	events := []alert.Event{makeNormalizeEvent("127.0.0.1", "TCP")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Entry.Proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", got[0].Entry.Proto)
	}
}

func TestNormalizeNotifier_UnspecifiedIPv4(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewNormalizeNotifier(inner, nil)
	events := []alert.Event{makeNormalizeEvent("0.0.0.0", "tcp")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Entry.Addr != "0.0.0.0" {
		t.Errorf("expected addr '0.0.0.0', got %q", got[0].Entry.Addr)
	}
}

func TestNormalizeNotifier_BlankAddressDefaultsToUnspecified(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewNormalizeNotifier(inner, nil)
	events := []alert.Event{makeNormalizeEvent("", "udp")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Entry.Addr != "0.0.0.0" {
		t.Errorf("expected addr '0.0.0.0', got %q", got[0].Entry.Addr)
	}
}

func TestNormalizeNotifier_CustomFn(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	customFn := func(e alert.Event) alert.Event {
		e.Entry.Addr = "custom"
		return e
	}
	n := alert.NewNormalizeNotifier(inner, customFn)
	events := []alert.Event{makeNormalizeEvent("192.168.1.1", "tcp")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Entry.Addr != "custom" {
		t.Errorf("expected addr 'custom', got %q", got[0].Entry.Addr)
	}
}
