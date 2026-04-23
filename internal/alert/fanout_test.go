package alert_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/your-org/portwatch/internal/alert"
	"github.com/your-org/portwatch/internal/scanner"
)

func makeFanoutEvent() alert.Event {
	return alert.Event{
		Kind: alert.KindAdded,
		Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9000, Protocol: "tcp"},
	}
}

func TestFanoutNotifier_EmptyEvents(t *testing.T) {
	var called int32
	n := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})
	f := alert.NewFanoutNotifier(n)
	if err := f.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatal("expected notifier not to be called on empty events")
	}
}

func TestFanoutNotifier_CallsAllNotifiers(t *testing.T) {
	events := []alert.Event{makeFanoutEvent()}
	var count int32
	make := func() alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
			atomic.AddInt32(&count, 1)
			return nil
		})
	}
	f := alert.NewFanoutNotifier(make(), make(), make())
	if err := f.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&count) != 3 {
		t.Fatalf("expected 3 calls, got %d", count)
	}
}

func TestFanoutNotifier_ReturnsErrorFromFailingNotifier(t *testing.T) {
	events := []alert.Event{makeFanoutEvent()}
	ok := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error { return nil })
	bad := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("downstream failure")
	})
	f := alert.NewFanoutNotifier(ok, bad)
	err := f.Send(context.Background(), events)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "downstream failure") {
		t.Errorf("expected error to contain 'downstream failure', got: %v", err)
	}
}

func TestFanoutNotifier_CollectsMultipleErrors(t *testing.T) {
	events := []alert.Event{makeFanoutEvent()}
	fail := func(msg string) alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
			return errors.New(msg)
		})
	}
	f := alert.NewFanoutNotifier(fail("err-alpha"), fail("err-beta"))
	err := f.Send(context.Background(), events)
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !strings.Contains(err.Error(), "err-alpha") || !strings.Contains(err.Error(), "err-beta") {
		t.Errorf("expected both errors in output, got: %v", err)
	}
}

func TestFanoutNotifier_NoNotifiers(t *testing.T) {
	f := alert.NewFanoutNotifier()
	if err := f.Send(context.Background(), []alert.Event{makeFanoutEvent()}); err != nil {
		t.Fatalf("unexpected error with no notifiers: %v", err)
	}
}
