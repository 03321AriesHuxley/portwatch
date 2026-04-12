package daemon_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/user/portwatch/internal/daemon"
)

fSignals_CancelledByParent(t *testing.T) {
	paCancel := context.WithCancelx, cancel := daemon	defer cancel()

rentCancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled after parent cancellation")
	}
}

func TestWithSignals_CancelledBySIGINT(t *testing.T) {
	ctx, cancel := daemon.WithSignals(context.Background())
	defer cancel()

	// Send SIGINT to ourselves.
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess: %v", err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("Signal: %v", err)
	}

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled after SIGINT")
	}
}

func TestWithSignals_ExplicitCancel(t *testing.T) {
	ctx, cancel := daemon.WithSignals(context.Background())
	cancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not cancelled after explicit cancel")
	}
}
