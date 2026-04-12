package scanner

import (
	"testing"
)

func makeEntry(proto, addr string, port uint16) Entry {
	return Entry{Proto: proto, LocalAddr: addr, LocalPort: port}
}

func TestDiff_NoDifference(t *testing.T) {
	prev := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
		makeEntry("tcp", "0.0.0.0", 443),
	}
	curr := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
		makeEntry("tcp", "0.0.0.0", 443),
	}

	changes := Diff(prev, curr)
	if len(changes) != 0 {
		t.Errorf("expected no changes, got %d", len(changes))
	}
}

func TestDiff_DetectsAddedPort(t *testing.T) {
	prev := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
	}
	curr := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
		makeEntry("tcp", "0.0.0.0", 8080),
	}

	changes := Diff(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != Added {
		t.Errorf("expected Added, got %s", changes[0].Type)
	}
	if changes[0].Entry.LocalPort != 8080 {
		t.Errorf("expected port 8080, got %d", changes[0].Entry.LocalPort)
	}
}

func TestDiff_DetectsRemovedPort(t *testing.T) {
	prev := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
		makeEntry("tcp", "0.0.0.0", 9090),
	}
	curr := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
	}

	changes := Diff(prev, curr)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != Removed {
		t.Errorf("expected Removed, got %s", changes[0].Type)
	}
	if changes[0].Entry.LocalPort != 9090 {
		t.Errorf("expected port 9090, got %d", changes[0].Entry.LocalPort)
	}
}

func TestDiff_EmptySnapshots(t *testing.T) {
	changes := Diff(nil, nil)
	if len(changes) != 0 {
		t.Errorf("expected no changes for empty snapshots, got %d", len(changes))
	}
}

func TestDiff_MultipleChanges(t *testing.T) {
	prev := []Entry{
		makeEntry("tcp", "0.0.0.0", 22),
		makeEntry("tcp", "0.0.0.0", 80),
	}
	curr := []Entry{
		makeEntry("tcp", "0.0.0.0", 80),
		makeEntry("tcp", "0.0.0.0", 443),
		makeEntry("tcp", "0.0.0.0", 8443),
	}

	changes := Diff(prev, curr)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	counts := map[ChangeType]int{}
	for _, c := range changes {
		counts[c.Type]++
	}
	if counts[Added] != 2 {
		t.Errorf("expected 2 added, got %d", counts[Added])
	}
	if counts[Removed] != 1 {
		t.Errorf("expected 1 removed, got %d", counts[Removed])
	}
}
