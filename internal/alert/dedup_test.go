package alert

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

func makeDedupEntry(proto, addr string, port uint16) scanner.Entry {
	return scanner.Entry{Proto: proto, LocalAddr: addr, LocalPort: port}
}

func TestDeduplicator_AllowsFirstOccurrence(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	events := []Event{
		{Kind: "added", Entry: makeDedupEntry("tcp", "0.0.0.0", 8080)},
	}
	out := d.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out))
	}
}

func TestDeduplicator_SuppressesDuplicate(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	events := []Event{
		{Kind: "added", Entry: makeDedupEntry("tcp", "0.0.0.0", 8080)},
	}
	d.Filter(events)
	out := d.Filter(events)
	if len(out) != 0 {
		t.Fatalf("expected 0 events after dedup, got %d", len(out))
	}
}

func TestDeduplicator_AllowsAfterWindowExpires(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(5 * time.Second)
	d.now = func() time.Time { return now }

	events := []Event{
		{Kind: "added", Entry: makeDedupEntry("tcp", "0.0.0.0", 9090)},
	}
	d.Filter(events)

	// Advance time beyond the window.
	d.now = func() time.Time { return now.Add(10 * time.Second) }
	out := d.Filter(events)
	if len(out) != 1 {
		t.Fatalf("expected 1 event after window expiry, got %d", len(out))
	}
}

func TestDeduplicator_DifferentKindsSameEntry(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	entry := makeDedupEntry("tcp", "127.0.0.1", 443)
	events := []Event{
		{Kind: "added", Entry: entry},
		{Kind: "removed", Entry: entry},
	}
	out := d.Filter(events)
	if len(out) != 2 {
		t.Fatalf("expected 2 events for different kinds, got %d", len(out))
	}
}

func TestDeduplicator_EmptyInput(t *testing.T) {
	d := NewDeduplicator(30 * time.Second)
	out := d.Filter(nil)
	if len(out) != 0 {
		t.Fatalf("expected 0 events for nil input, got %d", len(out))
	}
}
