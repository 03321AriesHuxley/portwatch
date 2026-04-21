package alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/portwatch/internal/alert"
)

// TestEscalate_WithPipeline verifies that EscalateNotifier composes correctly
// inside a pipeline, escalating to a secondary when the primary chain fails.
func TestEscalate_WithPipeline(t *testing.T) {
	primaryErr := errors.New("primary pipeline failed")

	var primaryReceived, secondaryReceived int

	primary := alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
		primaryReceived += len(evts)
		return primaryErr
	})

	secondary := alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
		secondaryReceived += len(evts)
		return nil
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 0, Target: secondary},
	)

	events := []alert.Event{
		{Kind: "added", Entry: makeEntry(9090, "tcp", "127.0.0.1")},
		{Kind: "removed", Entry: makeEntry(443, "tcp", "0.0.0.0")},
	}

	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("expected escalation to absorb error, got: %v", err)
	}

	if primaryReceived != 2 {
		t.Errorf("primary expected 2 events, got %d", primaryReceived)
	}
	if secondaryReceived != 2 {
		t.Errorf("secondary expected 2 events, got %d", secondaryReceived)
	}
}
