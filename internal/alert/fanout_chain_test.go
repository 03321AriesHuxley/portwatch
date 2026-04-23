package alert_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/your-org/portwatch/internal/alert"
	"github.com/your-org/portwatch/internal/scanner"
)

// TestFanoutThenBatch verifies that each fanout branch can independently batch events.
func TestFanoutThenBatch(t *testing.T) {
	var mu sync.Mutex
	var flushed [][]alert.Event

	recorder := alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
		mu.Lock()
		defer mu.Unlock()
		flushed = append(flushed, evts)
		return nil
	})

	batch := alert.NewBatchNotifier(recorder, 10, 50*time.Millisecond)
	fan := alert.NewFanoutNotifier(batch)

	events := []alert.Event{
		{Kind: alert.KindAdded, Entry: scanner.Entry{Port: 1111}},
		{Kind: alert.KindAdded, Entry: scanner.Entry{Port: 2222}},
	}

	if err := fan.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait for batch flush
	time.Sleep(100 * time.Millisecond)
	_ = batch.Flush(context.Background())

	mu.Lock()
	defer mu.Unlock()
	if len(flushed) == 0 {
		t.Fatal("expected at least one batch flush")
	}
}

// TestThrottleThenFanout verifies throttling upstream of fanout limits delivery to all branches.
func TestThrottleThenFanout(t *testing.T) {
	var mu sync.Mutex
	counts := map[string]int{}

	recorder := func(name string) alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
			mu.Lock()
			defer mu.Unlock()
			counts[name] += len(evts)
			return nil
		})
	}

	fan := alert.NewFanoutNotifier(recorder("a"), recorder("b"))
	throttle := alert.NewThrottleNotifier(fan, 1*time.Second)

	event := alert.Event{Kind: alert.KindAdded, Entry: scanner.Entry{Port: 5000}}

	_ = throttle.Send(context.Background(), []alert.Event{event})
	_ = throttle.Send(context.Background(), []alert.Event{event}) // should be throttled

	mu.Lock()
	defer mu.Unlock()
	for _, name := range []string{"a", "b"} {
		if counts[name] > 1 {
			t.Errorf("%s: throttle should have suppressed second send, got %d", name, counts[name])
		}
	}
}
