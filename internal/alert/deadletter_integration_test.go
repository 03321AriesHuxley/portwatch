package alert_test

import (
	"context"
	"errors"
	"testing"

	"portwatch/internal/alert"
	"portwatch/internal/scanner"
)

// TestDeadLetter_WithRetryNotifier verifies that events exhausting all retry
// attempts are captured by the dead-letter queue.
func TestDeadLetter_WithRetryNotifier(t *testing.T) {
	alwaysFail := &alwaysFailNotifier{err: errors.New("permanent failure")}

	// RetryNotifier wraps the failing notifier — it will exhaust all attempts.
	retryN := alert.NewRetryNotifier(alwaysFail, 3, 0)

	// DeadLetterNotifier wraps the retry notifier.
	dl := alert.NewDeadLetterNotifier(retryN, 10)

	events := []alert.Event{
		{
			Kind:  alert.KindAdded,
			Entry: scanner.Entry{Addr: "127.0.0.1", Port: 8080, Protocol: "tcp"},
		},
	}

	err := dl.Send(context.Background(), events)
	if err == nil {
		t.Fatal("expected an error from exhausted retries")
	}

	if dl.Len() != 1 {
		t.Fatalf("expected 1 dead-letter entry, got %d", dl.Len())
	}

	drained := dl.Drain()
	if len(drained) != 1 {
		t.Fatalf("expected 1 drained entry, got %d", len(drained))
	}
	if len(drained[0].Events) != 1 {
		t.Errorf("expected 1 event in dead-letter entry, got %d", len(drained[0].Events))
	}
	if dl.Len() != 0 {
		t.Errorf("expected empty queue after drain")
	}
}

// alwaysFailNotifier is a test double that always returns an error.
type alwaysFailNotifier struct{ err error }

func (a *alwaysFailNotifier) Send(_ context.Context, _ []alert.Event) error { return a.err }
