//go:build integration

package daemon_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/daemon"
)

// TestWatcher_Integration runs the watcher against a real proc-like file for
// a short duration and verifies it exits cleanly on context cancellation.
// Run with: go test -tags integration ./internal/daemon/...
func TestWatcher_Integration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tcp")

	initial := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n" +
		"   0: 0F02000A:0035 00000000:0000 0A 00000000:00000000 00:00000000 00000000   101        0 123\n"

	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	logger := alert.NewLogger(nil)
	watcher := daemon.NewWatcher(50*time.Millisecond, path, logger, nil)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- watcher.Run(ctx)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("watcher did not stop after context cancel")
	}
}
