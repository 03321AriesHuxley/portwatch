package alert_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeTeeEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{
			LocalAddress: "0.0.0.0",
			LocalPort: port,
			Protocol: "tcp",
		},
	}
}

func TestTeeNotifier_EmptyEvents(t *testing.T) {
	called := false
	n := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		called = true
		return nil
	})
	tee := alert.NewTeeNotifier(n)
	if err := tee.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected notifier not to be called for empty events")
	}
}

func TestTeeNotifier_CallsAllNotifiers(t *testing.T) {
	counts := make([]int, 3)
	makeN := func(i int) alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
			counts[i] += len(events)
			return nil
		})
	}
	tee := alert.NewTeeNotifier(makeN(0), makeN(1), makeN(2))
	events := []alert.Event{makeTeeEvent(8080), makeTeeEvent(9090)}
	if err := tee.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range counts {
		if c != 2 {
			t.Errorf("notifier %d: expected 2 events, got %d", i, c)
		}
	}
}

func TestTeeNotifier_ContinuesAfterError(t *testing.T) {
	secondCalled := false
	failing := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("boom")
	})
	successful := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		secondCalled = true
		return nil
	})
	tee := alert.NewTeeNotifier(failing, successful)
	err := tee.Send(context.Background(), []alert.Event{makeTeeEvent(443)})
	if err == nil {
		t.Fatal("expected error from failing notifier")
	}
	if !secondCalled {
		t.Fatal("expected second notifier to be called despite first failing")
	}
}

func TestTeeNotifier_CombinesMultipleErrors(t *testing.T) {
	makeErr := func(msg string) alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
			return errors.New(msg)
		})
	}
	tee := alert.NewTeeNotifier(makeErr("err1"), makeErr("err2"))
	err := tee.Send(context.Background(), []alert.Event{makeTeeEvent(80)})
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !strings.Contains(err.Error(), "err1") || !strings.Contains(err.Error(), "err2") {
		t.Errorf("expected both errors in message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "2 notifier(s) failed") {
		t.Errorf("expected failure count in message, got: %v", err)
	}
}
