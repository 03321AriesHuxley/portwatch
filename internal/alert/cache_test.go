package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeCacheEvent(port uint16) alert.Event {
	return alert.Event{
		Kind: "added",
		Entry: scanner.Entry{Port: port, Protocol: "tcp", Address: "0.0.0.0"},
	}
}

func TestCacheNotifier_StoresOnSuccess(t *testing.T) {
	var received []alert.Event
	inner := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = evs
		return nil
	})
	c := alert.NewCacheNotifier(inner, time.Minute)
	events := []alert.Event{makeCacheEvent(8080)}
	if err := c.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cached, at := c.Cached()
	if len(cached) != 1 {
		t.Fatalf("expected 1 cached event, got %d", len(cached))
	}
	if at.IsZero() {
		t.Error("expected non-zero cachedAt")
	}
	_ = received
}

func TestCacheNotifier_DoesNotUpdateOnFailure(t *testing.T) {
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("send failed")
	})
	c := alert.NewCacheNotifier(inner, time.Minute)
	_ = c.Send(context.Background(), []alert.Event{makeCacheEvent(9090)})
	cached, _ := c.Cached()
	if cached != nil {
		t.Error("expected nil cache after failure")
	}
}

func TestCacheNotifier_ExpiresTTL(t *testing.T) {
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error { return nil })
	c := alert.NewCacheNotifier(inner, 100*time.Millisecond)
	_ = c.Send(context.Background(), []alert.Event{makeCacheEvent(443)})
	time.Sleep(150 * time.Millisecond)
	cached, _ := c.Cached()
	if cached != nil {
		t.Error("expected cache to be expired")
	}
}

func TestCacheNotifier_Invalidate(t *testing.T) {
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error { return nil })
	c := alert.NewCacheNotifier(inner, time.Minute)
	_ = c.Send(context.Background(), []alert.Event{makeCacheEvent(22)})
	c.Invalidate()
	cached, _ := c.Cached()
	if cached != nil {
		t.Error("expected nil after Invalidate")
	}
}

func TestCacheNotifier_EmptyEventsSkipped(t *testing.T) {
	called := false
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		called = true
		return nil
	})
	c := alert.NewCacheNotifier(inner, time.Minute)
	if err := c.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("inner should not be called for empty events")
	}
}
