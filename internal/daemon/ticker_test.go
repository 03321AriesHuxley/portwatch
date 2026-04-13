package daemon

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTicker_FiresAtInterval(t *testing.T) {
	t.Parallel()

	tick := NewTicker(20 * time.Millisecond)
	defer tick.Stop()

	select {
	case <-tick.C():
		// received a tick as expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected tick did not arrive within timeout")
	}
}

func TestTicker_Stop(t *testing.T) {
	t.Parallel()

	tick := NewTicker(20 * time.Millisecond)
	tick.Stop()

	// Drain any tick already queued before Stop.
	time.Sleep(10 * time.Millisecond)
	for len(tick.C()) > 0 {
		<-tick.C()
	}

	select {
	case <-tick.C():
		t.Fatal("received tick after Stop")
	case <-time.After(60 * time.Millisecond):
		// correctly silent after stop
	}
}

func TestRunEvery_CallsFn(t *testing.T) {
	t.Parallel()

	var count atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	err := RunEvery(ctx, 30*time.Millisecond, func() {
		count.Add(1)
	})

	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}

	if n := count.Load(); n < 2 {
		t.Fatalf("expected fn to be called at least 2 times, got %d", n)
	}
}

func TestRunEvery_RespectsCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := RunEvery(ctx, 10*time.Millisecond, func() {
		t.Error("fn should not be called when context is already cancelled")
	})

	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
