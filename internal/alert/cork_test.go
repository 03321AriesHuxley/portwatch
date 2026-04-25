package alert_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeCorkEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{
			LocalAddress: "0.0.0.0",
			LocalPort:    port,
			Protocol:     "tcp",
		},
	}
}

func TestCorkNotifier_ForwardsWhenUncorked(t *testing.T) {
	var received []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	c := alert.NewCorkNotifier(next, false)
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(80)})

	if len(received) != 1 {
		t.Fatalf("expected 1 event forwarded, got %d", len(received))
	}
}

func TestCorkNotifier_BuffersWhenCorked(t *testing.T) {
	var received []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	c := alert.NewCorkNotifier(next, true)
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(443)})
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(8080)})

	if len(received) != 0 {
		t.Fatalf("expected 0 forwarded while corked, got %d", len(received))
	}
	if len(c.Buffered()) != 2 {
		t.Fatalf("expected 2 buffered, got %d", len(c.Buffered()))
	}
}

func TestCorkNotifier_FlushesOnUncork(t *testing.T) {
	var received []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	c := alert.NewCorkNotifier(next, true)
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(22)})
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(3306)})

	if err := c.Uncork(context.Background()); err != nil {
		t.Fatalf("unexpected error on Uncork: %v", err)
	}
	if len(received) != 2 {
		t.Fatalf("expected 2 events after uncork, got %d", len(received))
	}
	if len(c.Buffered()) != 0 {
		t.Fatalf("expected buffer empty after uncork, got %d", len(c.Buffered()))
	}
}

func TestCorkNotifier_DrainDiscardsBuffer(t *testing.T) {
	var received []alert.Event
	next := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	c := alert.NewCorkNotifier(next, true)
	_ = c.Send(context.Background(), []alert.Event{makeCorkEvent(9090)})
	c.Drain()

	_ = c.Uncork(context.Background())
	if len(received) != 0 {
		t.Fatalf("expected 0 events after drain+uncork, got %d", len(received))
	}
}

func TestCorkNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	next := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})

	c := alert.NewCorkNotifier(next, false)
	_ = c.Send(context.Background(), nil)

	if called {
		t.Fatal("expected inner notifier not to be called for empty events")
	}
}
