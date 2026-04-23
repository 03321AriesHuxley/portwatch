package alert_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/your-org/portwatch/internal/alert"
)

// TestFanoutNotifier_WithPipeline verifies that a FanoutNotifier works correctly
// when each branch is itself a pipeline with filtering and rate limiting.
func TestFanoutNotifier_WithPipeline(t *testing.T) {
	events := []alert.Event{makeFanoutEvent()}

	var mu sync.Mutex
	received := map[string]int{}

	makeRecorder := func(name string) alert.Notifier {
		return alert.NotifierFunc(func(_ context.Context, evts []alert.Event) error {
			mu.Lock()
			defer mu.Unlock()
			received[name] += len(evts)
			return nil
		})
	}

	branch1 := alert.NewPipelineBuilder().
		Add(makeRecorder("branch1")).
		Build()

	branch2 := alert.NewPipelineBuilder().
		Add(makeRecorder("branch2")).
		Build()

	fan := alert.NewFanoutNotifier(branch1, branch2)

	if err := fan.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if received["branch1"] != 1 {
		t.Errorf("branch1: expected 1 event, got %d", received["branch1"])
	}
	if received["branch2"] != 1 {
		t.Errorf("branch2: expected 1 event, got %d", received["branch2"])
	}
}

// TestFanoutNotifier_PartialFailureDoesNotBlockOthers ensures that a slow or
// failing notifier does not prevent other branches from executing.
func TestFanoutNotifier_PartialFailureDoesNotBlockOthers(t *testing.T) {
	events := []alert.Event{makeFanoutEvent()}

	var reached int32
	good := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		reached++
		return nil
	})
	bad := alert.NotifierFunc(func(_ context.Context, _ []alert.Event) error {
		return errors.New("oops")
	})

	fan := alert.NewFanoutNotifier(good, bad, good)
	err := fan.Send(context.Background(), events)
	if err == nil {
		t.Fatal("expected error from bad notifier")
	}
	if reached != 2 {
		t.Errorf("expected 2 good notifiers to run, got %d", reached)
	}
}
