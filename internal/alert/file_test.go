package alert_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeFileEntry(proto, addr string, port uint16) scanner.Entry {
	return scanner.Entry{Protocol: proto, LocalAddress: addr, LocalPort: port}
}

func TestFileNotifier_EmptyEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "portwatch.log")
	n := alert.NewFileNotifier(path, "")
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := os.Stat(path)
	if err == nil {
		t.Error("expected no file to be created for empty events")
	}
}

func TestFileNotifier_CreatesFileAndWritesEvent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "portwatch.log")
	n := alert.NewFileNotifier(path, "[pw]")
	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeFileEntry("tcp", "0.0.0.0", 9000)},
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read log file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "[pw]") {
		t.Errorf("expected prefix [pw] in file, got: %q", content)
	}
	if !strings.Contains(content, "9000") {
		t.Errorf("expected port 9000 in file, got: %q", content)
	}
}

func TestFileNotifier_AppendsMultipleSends(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "portwatch.log")
	n := alert.NewFileNotifier(path, "")

	batch1 := []alert.Event{{Kind: alert.EventAdded, Entry: makeFileEntry("tcp", "0.0.0.0", 80)}}
	batch2 := []alert.Event{{Kind: alert.EventRemoved, Entry: makeFileEntry("tcp", "0.0.0.0", 80)}}

	if err := n.Send(context.Background(), batch1); err != nil {
		t.Fatalf("send 1 error: %v", err)
	}
	if err := n.Send(context.Background(), batch2); err != nil {
		t.Fatalf("send 2 error: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines after two sends, got %d", len(lines))
	}
}

func TestFileNotifier_InvalidPath(t *testing.T) {
	n := alert.NewFileNotifier("/nonexistent/dir/portwatch.log", "")
	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeFileEntry("tcp", "0.0.0.0", 22)},
	}
	if err := n.Send(context.Background(), events); err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}
