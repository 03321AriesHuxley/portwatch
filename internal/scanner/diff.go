package scanner

import "fmt"

// ChangeType indicates whether a port was added or removed.
type ChangeType string

const (
	Added   ChangeType = "added"
	Removed ChangeType = "removed"
)

// Change represents a single port binding change.
type Change struct {
	Type  ChangeType
	Entry Entry
}

func (c Change) String() string {
	return fmt.Sprintf("%s: %s", c.Type, c.Entry)
}

// Diff computes the difference between two snapshots, returning a list
// of changes (added or removed port entries).
func Diff(prev, curr []Entry) []Change {
	prevSet := toSet(prev)
	currSet := toSet(curr)

	var changes []Change

	for key, entry := range currSet {
		if _, exists := prevSet[key]; !exists {
			changes = append(changes, Change{Type: Added, Entry: entry})
		}
	}

	for key, entry := range prevSet {
		if _, exists := currSet[key]; !exists {
			changes = append(changes, Change{Type: Removed, Entry: entry})
		}
	}

	return changes
}

// toSet converts a slice of Entry into a map keyed by a unique string
// representation for fast lookup.
func toSet(entries []Entry) map[string]Entry {
	m := make(map[string]Entry, len(entries))
	for _, e := range entries {
		m[e.String()] = e
	}
	return m
}
