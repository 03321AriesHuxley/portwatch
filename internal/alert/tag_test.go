package alert_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeTagEvent(kind alert.EventKind, port uint16, proto string) alert.Event {
	return alert.Event{
		Kind: kind,
		Entry: scanner.Entry{
			Port:     port,
			Address:  "0.0.0.0",
			Protocol: proto,
		},
	}
}

func TestTagNotifier_EmptyEvents(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTagNotifier(inner)
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 events, got %d", len(got))
	}
}

func TestTagNotifier_DefaultTags(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTagNotifier(inner)
	events := []alert.Event{makeTagEvent(alert.EventAdded, 8080, "tcp")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	tags := got[0].Meta["tags"]
	if tags == "" {
		t.Error("expected tags to be set")
	}
	if tags != "added,tcp" {
		t.Errorf("unexpected tags %q", tags)
	}
}

func TestTagNotifier_CustomTagger(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	custom := alert.WithTagger(func(e alert.Event) []string {
		return []string{"custom", "env:prod"}
	})
	n := alert.NewTagNotifier(inner, custom)
	events := []alert.Event{makeTagEvent(alert.EventAdded, 443, "tcp")}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["tags"] != "custom,env:prod" {
		t.Errorf("unexpected tags %q", got[0].Meta["tags"])
	}
}

func TestTagNotifier_PreservesExistingMeta(t *testing.T) {
	var got []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		got = events
		return nil
	})
	n := alert.NewTagNotifier(inner)
	e := makeTagEvent(alert.EventRemoved, 22, "tcp")
	e.Meta = map[string]string{"source": "scanner"}
	if err := n.Send(context.Background(), []alert.Event{e}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].Meta["source"] != "scanner" {
		t.Errorf("existing meta key lost, got %v", got[0].Meta)
	}
	if got[0].Meta["tags"] == "" {
		t.Error("tags not set")
	}
}
