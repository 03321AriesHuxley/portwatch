package alert_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

// TestCacheNotifier_WithRetry verifies that CacheNotifier preserves a
// previous successful batch while a transient failure is being retried.
func TestCacheNotifier_WithRetry(t *testing.T) {
	var calls atomic.Int32

	// Fail the first attempt, succeed on the second.
	inner := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		if calls.Add(1) == 1 {
			return errors.New("transient")
		}
		return nil
	})

	retry := alert.NewRetryNotifier(inner, 3, 5*time.Millisecond)
	cache := alert.NewCacheNotifier(retry, time.Minute)

	events := []alert.Event{
		{Kind: "added", Entry: scanner.Entry{Port: 8443, Protocol: "tcp", Address: "0.0.0.0"}},
	}

	if err := cache.Send(context.Background(), events); err != nil {
		t.Fatalf("expected eventual success, got: %v", err)
	}

	cached, at := cache.Cached()
	if len(cached) != 1 {
		t.Fatalf("expected 1 cached event, got %d", len(cached))
	}
	if at.IsZero() {
		t.Error("expected non-zero cachedAt")
	}
	if calls.Load() < 2 {
		t.Errorf("expected at least 2 inner calls (retry), got %d", calls.Load())
	}
}
