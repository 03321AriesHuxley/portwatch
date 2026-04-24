package alert_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeDebounceEvents(port uint16) []alert.Event {
	return []alert.Event{
		{
			Kind: alert.EventAdded,
			Entry: scanner.Entry{
				LocalAddress: "0.0.0.0",
				LocalPort:    port,
				Protocol:     "tcp",
			},
		},
	}
}

func TestDebounceNotifier_ForwardsAfterQuietWindow(t *testing.T) {
	var called int32
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	d := alert.NewDebounceNotifier(inner, 30*time.Millisecond)
	err := d.Send(context.Background(), makeDebounceEvents(9090))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected inner called once, got %d", called)
	}
}

func TestDebounceNotifier_EmptyEventsSkipped(t *testing.T) {
	var called int32
	inner := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		atomic.AddInt32(&called, 1)
		return nil
	})

	d := alert.NewDebounceNotifier(inner, 30*time.Millisecond)
	err := d.Send(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&called) != 0 {
		t.Fatalf("expected inner not called, got %d", called)
	}
}

func TestDebounceNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return sentinel
	})

	d := alert.NewDebounceNotifier(inner, 20*time.Millisecond)
	err := d.Send(context.Background(), makeDebounceEvents(8080))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDebounceNotifier_RespectsContextCancel(t *testing.T) {
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	d := alert.NewDebounceNotifier(inner, 150*time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := d.Send(ctx, makeDebounceEvents(7070))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}
