package scanner

import "fmt"

// ChangeKind describes whether a port was added or removed.
type ChangeKind string

const (
	Added   ChangeKind = "added"
	Removed ChangeKind = "removed"
)

// Change represents a single port binding change between two snapshots.
type Change struct {
	Kind  ChangeKind
	Entry PortEntry
}

func (c Change) String() string {
	return fmt.Sprintf("[%s] %s port %d (addr %s)",
		c.Kind, c.Entry.Protocol, c.Entry.Port, c.Entry.LocalAddr)
}

// Diff compares two snapshots and returns the list of changes.
// prev is the previous snapshot; curr is the current one.
func Diff(prev, curr []PortEntry) []Change {
	prevSet := toSet(prev)
	currSet := toSet(curr)

	var changes []Change

	for key, entry := range currSet {
		if _, exists := prevSet[key]; !exists {
			changes = append(changes, Change{Kind: Added, Entry: entry})
		}
	}
	for key, entry := range prevSet {
		if _, exists := currSet[key]; !exists {
			changes = append(changes, Change{Kind: Removed, Entry: entry})
		}
	}
	return changes
}

func toSet(entries []PortEntry) map[string]PortEntry {
	m := make(map[string]PortEntry, len(entries))
	for _, e := range entries {
		key := fmt.Sprintf("%s|%s|%d", e.Protocol, e.LocalAddr, e.Port)
		m[key] = e
	}
	return m
}
