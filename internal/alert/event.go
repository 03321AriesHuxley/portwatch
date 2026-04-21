package alert

import (
	"github.com/example/portwatch/internal/scanner"
)

// EventKind describes whether a port was added or removed.
type EventKind string

const (
	EventAdded   EventKind = "added"
	EventRemoved EventKind = "removed"
)

// Event represents a single port binding change detected by the scanner.
type Event struct {
	Kind  EventKind
	Entry scanner.Entry
	// Meta holds arbitrary key-value annotations attached by enrichers or taggers.
	Meta map[string]string
}

// EventsFromDiff converts a scanner.Diff into a slice of Events.
func EventsFromDiff(d scanner.Diff) []Event {
	events := make([]Event, 0, len(d.Added)+len(d.Removed))
	for _, e := range d.Added {
		events = append(events, Event{Kind: EventAdded, Entry: e})
	}
	for _, e := range d.Removed {
		events = append(events, Event{Kind: EventRemoved, Entry: e})
	}
	return events
}
