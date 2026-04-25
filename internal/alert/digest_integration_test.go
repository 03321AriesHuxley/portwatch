package alert_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/patrickdappollonio/portwatch/internal/alert"
)

func TestDigestNotifier_Integration(t *testing.T) {
	var mu sync.Mutex
	var received []alert.Event

	sink := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, events...)
		return nil
	})

	now := time.Now()
	d := alert.NewDigestNotifier(sink, 100*time.Millisecond)
	d.SetClock(func() time.Time { return now })

	evt := alert.Event{Kind: "added"}

	// Send three events — all within the same window.
	for i := 0; i < 3; i++ {
		if err := d.Send(context.Background(), []alert.Event{evt}); err != nil {
			t.Fatalf("unexpected send error: %v", err)
		}
	}

	mu.Lock()
	beforeFlush := len(received)
	mu.Unlock()

	if beforeFlush != 0 {
		t.Fatalf("expected no events before flush, got %d", beforeFlush)
	}

	if err := d.Flush(context.Background()); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	mu.Lock()
	afterFlush := len(received)
	mu.Unlock()

	if afterFlush != 3 {
		t.Fatalf("expected 3 events after flush, got %d", afterFlush)
	}
}
