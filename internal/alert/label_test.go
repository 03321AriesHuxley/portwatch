package alert_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeLabelEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{
			Address:  "0.0.0.0",
			Port:     port,
			Protocol: "tcp",
		},
		Meta: map[string]string{},
	}
}

func TestLabelNotifier_EmptyEvents(t *testing.T) {
	var called bool
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	n := alert.NewLabelNotifier(next, alert.WithLabeler(map[string]string{"env": "test"}))
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected next not to be called for empty events")
	}
}

func TestLabelNotifier_AttachesStaticLabels(t *testing.T) {
	var got []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	labels := map[string]string{"env": "prod", "region": "us-east-1"}
	n := alert.NewLabelNotifier(next, alert.WithLabeler(labels))

	events := []alert.Event{makeLabelEvent(8080)}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	if got[0].Meta["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got[0].Meta["env"])
	}
	if got[0].Meta["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", got[0].Meta["region"])
	}
}

func TestLabelNotifier_ExistingMetaNotOverwritten(t *testing.T) {
	var got []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewLabelNotifier(next, alert.WithLabeler(map[string]string{"env": "prod"}))

	e := makeLabelEvent(9090)
	e.Meta["env"] = "staging" // pre-existing value should win

	if err := n.Send(context.Background(), []alert.Event{e}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["env"] != "staging" {
		t.Errorf("expected existing meta to win, got %q", got[0].Meta["env"])
	}
}

func TestLabelNotifier_CustomLabelerPerEvent(t *testing.T) {
	var got []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	dynamic := func(e alert.Event) map[string]string {
		if e.Entry.Port == 443 {
			return map[string]string{"secure": "true"}
		}
		return map[string]string{"secure": "false"}
	}
	n := alert.NewLabelNotifier(next, dynamic)

	events := []alert.Event{makeLabelEvent(443), makeLabelEvent(80)}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["secure"] != "true" {
		t.Errorf("port 443 should be secure=true")
	}
	if got[1].Meta["secure"] != "false" {
		t.Errorf("port 80 should be secure=false")
	}
}
