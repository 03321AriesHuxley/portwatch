package alert_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// TestDebounce_InPipeline verifies that a DebounceNotifier placed inside a
// Pipeline correctly coalesces events before forwarding downstream.
func TestDebounce_InPipeline(t *testing.T) {
	var received int32
	sink := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		atomic.AddInt32(&received, int32(len(events)))
		return nil
	})

	pipeline := alert.NewPipeline(
		alert.NewDebounceNotifier(sink, 20*time.Millisecond),
	)

	events := []alert.Event{
		{
			Kind: alert.EventAdded,
			Entry: scanner.Entry{LocalAddress: "127.0.0.1", LocalPort: 3000, Protocol: "tcp"},
		},
		{
			Kind: alert.EventRemoved,
			Entry: scanner.Entry{LocalAddress: "127.0.0.1", LocalPort: 3001, Protocol: "tcp"},
		},
	}

	if err := pipeline.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&received); got != int32(len(events)) {
		t.Fatalf("expected %d events received, got %d", len(events), got)
	}
}

// TestDebounce_WithFilter verifies that events filtered before the debounce
// stage are never forwarded to the sink.
func TestDebounce_WithFilter(t *testing.T) {
	var received int32
	sink := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		atomic.AddInt32(&received, int32(len(events)))
		return nil
	})

	filtered := alert.NewFilter(alert.NewDebounceNotifier(sink, 20*time.Millisecond),
		alert.WithExcludedPorts(9999),
	)

	events := []alert.Event{
		{
			Kind:  alert.EventAdded,
			Entry: scanner.Entry{LocalAddress: "0.0.0.0", LocalPort: 9999, Protocol: "tcp"},
		},
	}

	if err := filtered.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := atomic.LoadInt32(&received); got != 0 {
		t.Fatalf("expected 0 events after filter, got %d", got)
	}
}
