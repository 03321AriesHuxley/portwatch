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

func writeProcNet(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "tcp")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeProcNet: %v", err)
	}
	return p
}

const procNetContent = `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 0 1 0000000000000000 100 0 0 10 0
`

func TestRunner_CancelImmediately(t *testing.T) {
	dir := t.TempDir()
	path := writeProcNet(t, dir, procNetContent)

	logger := alert.NewLogger(nil)
	r := daemon.NewRunner(100*time.Millisecond, path, logger, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run

	err := r.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRunner_DetectsChange(t *testing.T) {
	dir := t.TempDir()
	path := writeProcNet(t, dir, procNetContent)

	logger := alert.NewLogger(nil)
	r := daemon.NewRunner(30*time.Millisecond, path, logger, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// After a short delay, update the file to simulate a new port opening.
	go func() {
		time.Sleep(60 * time.Millisecond)
		newContent := procNetContent +
			"   1: 0100007F:1F91 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 0 1 0000000000000000 100 0 0 10 0\n"
		_ = os.WriteFile(path, []byte(newContent), 0o644)
	}()

	err := r.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestRunner_MissingProcFile(t *testing.T) {
	logger := alert.NewLogger(nil)
	r := daemon.NewRunner(50*time.Millisecond, "/nonexistent/tcp", logger, nil)

	ctx := context.Background()
	err := r.Run(ctx)
	if err == nil {
		t.Fatal("expected error for missing proc file")
	}
}
