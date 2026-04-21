package daemon

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func writeProcNetWatcher(t *testing.T, dir, content string) string {
	t.Helper()
	p := filepath.Join(dir, "tcp")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeProcNetWatcher: %v", err)
	}
	return p
}

const procHeader = "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n"

func TestWatcher_NoChanges(t *testing.T) {
	dir := t.TempDir()
	content := procHeader + "   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 0\n"
	path := writeProcNetWatcher(t, dir, content)

	logger := alert.NewLogger(nil)
	watcher := NewWatcher(20*time.Millisecond, path, logger, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	err := watcher.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWatcher_DetectsAddedPort(t *testing.T) {
	dir := t.TempDir()
	path := writeProcNetWatcher(t, dir, procHeader)

	hookCalled := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hookCalled <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hook, err := alert.NewHook(ts.URL, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHook: %v", err)
	}

	logger := alert.NewLogger(nil)
	watcher := NewWatcher(20*time.Millisecond, path, logger, hook)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watcher.Run(ctx) //nolint:errcheck

	time.Sleep(30 * time.Millisecond)
	newContent := procHeader + fmt.Sprintf("   0: 0100007F:1F91 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 0\n")
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-hookCalled:
		// success
	case <-time.After(300 * time.Millisecond):
		t.Fatal("hook was not called after port change")
	}
}

func TestWatcher_DetectsRemovedPort(t *testing.T) {
	dir := t.TempDir()
	initialContent := procHeader + "   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 0\n"
	path := writeProcNetWatcher(t, dir, initialContent)

	hookCalled := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hookCalled <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	hook, err := alert.NewHook(ts.URL, 2*time.Second)
	if err != nil {
		t.Fatalf("NewHook: %v", err)
	}

	logger := alert.NewLogger(nil)
	watcher := NewWatcher(20*time.Millisecond, path, logger, hook)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watcher.Run(ctx) //nolint:errcheck

	time.Sleep(30 * time.Millisecond)
	// Remove the port by writing only the header
	if err := os.WriteFile(path, []byte(procHeader), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	select {
	case <-hookCalled:
		// success
	case <-time.After(300 * time.Millisecond):
		t.Fatal("hook was not called after port removal")
	}
}

func TestWatcher_MissingProcFile(t *testing.T) {
	logger := alert.NewLogger(nil)
	watcher := NewWatcher(20*time.Millisecond, "/nonexistent/tcp", logger, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	err := watcher.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}
