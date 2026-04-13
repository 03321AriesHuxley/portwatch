package alert_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

type capturingNotifier struct {
	calls  atomic.Int32
	events []alert.Event
}

func (c *capturingNotifier) Send(_ context.Context, events []alert.Event) error {
	c.calls.Add(1)
	c.events = append(c.events, events...)
	return nil
}

func makeBatchEvent(kind alert.EventKind) alert.Event {
	return alert.Event{Kind: kind}
}

func TestBatchNotifier_FlushesOnMaxSize(t *testing.T) {
	inner := &capturingNotifier{}
	b := alert.NewBatchNotifier(inner, 5*time.Second, 3)
	ctx := context.Background()

	_ = b.Send(ctx, []alert.Event{makeBatchEvent(alert.EventAdded)})
	_ = b.Send(ctx, []alert.Event{makeBatchEvent(alert.EventAdded)})
	if inner.calls.Load() != 0 {
		t.Fatal("expected no flush before maxSize")
	}
	_ = b.Send(ctx, []alert.Event{makeBatchEvent(alert.EventRemoved)})
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 flush at maxSize, got %d", inner.calls.Load())
	}
	if len(inner.events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(inner.events))
	}
}

func TestBatchNotifier_FlushesAfterWindow(t *testing.T) {
	inner := &capturingNotifier{}
	b := alert.NewBatchNotifier(inner, 40*time.Millisecond, 100)
	ctx := context.Background()

	_ = b.Send(ctx, []alert.Event{makeBatchEvent(alert.EventAdded)})
	time.Sleep(80 * time.Millisecond)

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 flush after window, got %d", inner.calls.Load())
	}
}

func TestBatchNotifier_ManualFlush(t *testing.T) {
	inner := &capturingNotifier{}
	b := alert.NewBatchNotifier(inner, 5*time.Second, 100)
	ctx := context.Background()

	_ = b.Send(ctx, []alert.Event{makeBatchEvent(alert.EventAdded)})
	_ = b.Flush(ctx)

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call after manual flush, got %d", inner.calls.Load())
	}
}

func TestBatchNotifier_EmptyFlushIsNoop(t *testing.T) {
	inner := &capturingNotifier{}
	b := alert.NewBatchNotifier(inner, 5*time.Second, 10)

	_ = b.Flush(context.Background())
	if inner.calls.Load() != 0 {
		t.Fatal("expected no call on empty flush")
	}
}
