package alert_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/scanner"
)

func makeCorkPipelineEvent(port uint16, kind alert.Kind) alert.Event {
	return alert.Event{
		Kind: kind,
		Entry: scanner.Entry{
			LocalAddress: "127.0.0.1",
			LocalPort:    port,
			Protocol:     "tcp",
		},
	}
}

// TestCork_InPipeline verifies that a CorkNotifier placed inside a Pipeline
// buffers events while corked and forwards them after uncorking.
func TestCork_InPipeline(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	cork := alert.NewCorkNotifier(sink, true)

	p := alert.NewPipeline(cork)

	events := []alert.Event{
		makeCorkPipelineEvent(80, alert.KindAdded),
		makeCorkPipelineEvent(443, alert.KindAdded),
	}

	_ = p.Send(context.Background(), events)

	if len(received) != 0 {
		t.Fatalf("expected 0 forwarded while corked, got %d", len(received))
	}

	if err := cork.Uncork(context.Background()); err != nil {
		t.Fatalf("uncork error: %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 events after uncork, got %d", len(received))
	}
}

// TestCork_WithFilter verifies that events filtered out before the cork are
// never buffered.
func TestCork_WithFilter(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, events []alert.Event) error {
		received = append(received, events...)
		return nil
	})

	cork := alert.NewCorkNotifier(sink, true)
	filtered := alert.NewFilter(cork,
		alert.WithExcludedPorts(22),
	)

	events := []alert.Event{
		makeCorkPipelineEvent(22, alert.KindAdded),
		makeCorkPipelineEvent(8080, alert.KindAdded),
	}

	_ = filtered.Send(context.Background(), events)

	if len(cork.Buffered()) != 1 {
		t.Fatalf("expected 1 buffered (port 22 filtered), got %d", len(cork.Buffered()))
	}

	_ = cork.Uncork(context.Background())

	if len(received) != 1 {
		t.Fatalf("expected 1 received after uncork, got %d", len(received))
	}
	if received[0].Entry.LocalPort != 8080 {
		t.Fatalf("expected port 8080, got %d", received[0].Entry.LocalPort)
	}
}
