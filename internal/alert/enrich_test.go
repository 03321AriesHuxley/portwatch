package alert_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeEnrichEvent(port uint16, proto string) alert.Event {
	return alert.Event{
		Kind: alert.EventAdded,
		Entry: scanner.Entry{
			Port:     port,
			Address:  "127.0.0.1",
			Protocol: proto,
		},
	}
}

func TestEnrichNotifier_EmptyEvents(t *testing.T) {
	var called bool
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	n := alert.NewEnrichNotifier(inner, alert.StaticEnricher{Fields: map[string]string{"env": "test"}})
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("inner notifier not called")
	}
}

func TestEnrichNotifier_StaticEnricher(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	enr := alert.StaticEnricher{Fields: map[string]string{"env": "prod", "host": "box1"}}
	n := alert.NewEnrichNotifier(inner, enr)
	if err := n.Send(context.Background(), []alert.Event{makeEnrichEvent(80, "tcp")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got[0].Meta["env"])
	}
	if got[0].Meta["host"] != "box1" {
		t.Errorf("expected host=box1, got %q", got[0].Meta["host"])
	}
}

func TestEnrichNotifier_PortEnricher(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewEnrichNotifier(inner, alert.PortEnricher{})
	if err := n.Send(context.Background(), []alert.Event{makeEnrichEvent(443, "tcp")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["port_label"] != "tcp/443" {
		t.Errorf("unexpected port_label %q", got[0].Meta["port_label"])
	}
}

func TestEnrichNotifier_MultipleEnrichers(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewEnrichNotifier(inner,
		alert.StaticEnricher{Fields: map[string]string{"env": "staging"}},
		alert.PortEnricher{},
	)
	if err := n.Send(context.Background(), []alert.Event{makeEnrichEvent(22, "tcp")}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["env"] != "staging" {
		t.Errorf("missing env field")
	}
	if got[0].Meta["port_label"] != "tcp/22" {
		t.Errorf("missing port_label field")
	}
}
