package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeCachePipelineEvent(port uint16, kind string) alert.Event {
	return alert.Event{
		Kind:  kind,
		Entry: scanner.Entry{Port: port, Protocol: "tcp", Address: "127.0.0.1"},
	}
}

// TestCache_InPipeline verifies CacheNotifier integrates cleanly as the
// last stage of a pipeline, preserving the batch after a successful send.
func TestCache_InPipeline(t *testing.T) {
	var sent []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		sent = append(sent, evs...)
		return nil
	})

	cache := alert.NewCacheNotifier(sink, time.Minute)

	p := alert.NewPipelineBuilder().
		Add(alert.NewThrottleNotifier(cache, time.Millisecond, 100)).
		Build()

	events := []alert.Event{
		makeCachePipelineEvent(80, "added"),
		makeCachePipelineEvent(443, "added"),
	}

	if err := p.Send(context.Background(), events); err != nil {
		t.Fatalf("pipeline send failed: %v", err)
	}

	if len(sent) != 2 {
		t.Fatalf("expected 2 sent events, got %d", len(sent))
	}
}

// TestCache_PreservesOnPipelineFailure verifies the cache is not updated
// when the inner pipeline returns an error.
func TestCache_PreservesOnPipelineFailure(t *testing.T) {
	failSink := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("downstream failure")
	})

	cache := alert.NewCacheNotifier(failSink, time.Minute)

	_ = cache.Send(context.Background(), []alert.Event{makeCachePipelineEvent(22, "removed")})

	cached, _ := cache.Cached()
	if cached != nil {
		t.Error("cache should remain empty after failure")
	}
}
