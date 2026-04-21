package alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/yourusername/portwatch/internal/alert"
)

func TestEscalate_ChainedPolicies(t *testing.T) {
	primaryErr := errors.New("primary down")
	firstErr := errors.New("first escalation down")

	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return primaryErr
	})
	first := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return firstErr
	})

	var finalReceived int
	final := alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
		finalReceived += len(evts)
		return nil
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 0, Target: first},
		alert.EscalationPolicy{After: 0, Target: final},
	)

	events := []alert.Event{
		{Kind: "added", Entry: makeEntry(8443, "tcp", "0.0.0.0")},
	}

	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("expected final escalation to succeed, got: %v", err)
	}
	if finalReceived != 1 {
		t.Errorf("expected final notifier to receive 1 event, got %d", finalReceived)
	}
}

func TestEscalate_AllPoliciesFail_ReturnsOriginalError(t *testing.T) {
	origErr := errors.New("original")
	primary := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return origErr
	})
	fallback := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("fallback also failed")
	})

	n := alert.NewEscalateNotifier(primary,
		alert.EscalationPolicy{After: 0, Target: fallback},
	)

	err := n.Send(context.Background(), makeEscalateEvents())
	if err == nil {
		t.Fatal("expected error when all escalations fail")
	}
	if err.Error() != origErr.Error() {
		t.Errorf("expected original error %q, got %q", origErr, err)
	}
}
