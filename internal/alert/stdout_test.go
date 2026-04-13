package alert_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeStdoutEntry(proto, addr string, port uint16) scanner.Entry {
	return scanner.Entry{Protocol: proto, LocalAddress: addr, LocalPort: port}
}

func TestStdoutNotifier_EmptyEvents(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewStdoutNotifier(alert.WithWriter(&buf))
	if err := n.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestStdoutNotifier_WritesAddedEvent(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewStdoutNotifier(alert.WithWriter(&buf), alert.WithPrefix("[test]"))
	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeStdoutEntry("tcp", "0.0.0.0", 8080)},
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[test]") {
		t.Errorf("expected prefix [test] in output, got: %q", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %q", out)
	}
	if !strings.Contains(out, "added") {
		t.Errorf("expected 'added' in output, got: %q", out)
	}
}

func TestStdoutNotifier_WritesRemovedEvent(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewStdoutNotifier(alert.WithWriter(&buf))
	events := []alert.Event{
		{Kind: alert.EventRemoved, Entry: makeStdoutEntry("udp", "127.0.0.1", 5353)},
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "5353") {
		t.Errorf("expected port 5353 in output, got: %q", out)
	}
	if !strings.Contains(out, "removed") {
		t.Errorf("expected 'removed' in output, got: %q", out)
	}
}

func TestStdoutNotifier_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewStdoutNotifier(alert.WithWriter(&buf))
	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeStdoutEntry("tcp", "0.0.0.0", 80)},
		{Kind: alert.EventAdded, Entry: makeStdoutEntry("tcp", "0.0.0.0", 443)},
		{Kind: alert.EventRemoved, Entry: makeStdoutEntry("tcp", "0.0.0.0", 8080)},
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), buf.String())
	}
}

// TestStdoutNotifier_IncludesProtocolAndAddress verifies that the notifier
// includes the protocol and local address in its output for each event.
func TestStdoutNotifier_IncludesProtocolAndAddress(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewStdoutNotifier(alert.WithWriter(&buf))
	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: makeStdoutEntry("udp", "192.168.1.1", 9090)},
	}
	if err := n.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "udp") {
		t.Errorf("expected protocol 'udp' in output, got: %q", out)
	}
	if !strings.Contains(out, "192.168.1.1") {
		t.Errorf("expected address '192.168.1.1' in output, got: %q", out)
	}
}
