package alert_test

import (
	"context"
	"sync"
	"testing"

	"github.com/your-org/portwatch/internal/alert"
	"github.com/your-org/portwatch/internal/scanner"
)

// TestFanoutInPipeline verifies that a FanoutNotifier can be embedded as a
// stage inside a larger pipeline.
func TestFanoutInPipeline(t *testing.T) {
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

	pipeline := alert.NewPipelineBuilder().
		Add(alert.NewFanoutNotifier(
			recorder("sink-a"),
			recorder("sink-b"),
		)).
		Build()

	events := []alert.Event{
		{Kind: alert.KindAdded, Entry: scanner.Entry{Addr: "127.0.0.1", Port: 8080, Protocol: "tcp"}},
		{Kind: alert.KindRemoved, Entry: scanner.Entry{Addr: "0.0.0.0", Port: 443, Protocol: "tcp"}},
	}

	if err := pipeline.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, name := range []string{"sink-a", "sink-b"} {
		if counts[name] != 2 {
			t.Errorf("%s: expected 2 events, got %d", name, counts[name])
		}
	}
}

// TestFanoutWithFilter ensures filtering upstream of a fanout correctly
// reduces events delivered to all branches.
func TestFanoutWithFilter(t *testing.T) {
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

	// Exclude port 443 before fanout
	filter := alert.NewFilter(alert.FilterConfig{ExcludePorts: []uint16{443}})

	pipeline := alert.NewPipelineBuilder().
		Add(filter).
		Add(alert.NewFanoutNotifier(
			recorder("x"),
			recorder("y"),
		)).
		Build()

	events := []alert.Event{
		{Kind: alert.KindAdded, Entry: scanner.Entry{Addr: "0.0.0.0", Port: 8080, Protocol: "tcp"}},
		{Kind: alert.KindAdded, Entry: scanner.Entry{Addr: "0.0.0.0", Port: 443, Protocol: "tcp"}},
	}

	if err := pipeline.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	for _, name := range []string{"x", "y"} {
		if counts[name] != 1 {
			t.Errorf("%s: expected 1 event after filter, got %d", name, counts[name])
		}
	}
}
