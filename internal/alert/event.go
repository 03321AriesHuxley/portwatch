package alert

import (
	"fmt"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

// EventType describes whether a port was opened or closed.
type EventType string

const (
	EventOpened EventType = "opened"
	EventClosed EventType = "closed"
)

// PortEvent represents a single port binding change detected by the daemon.
type PortEvent struct {
	Type      EventType    `json:"type"`
	Entry     scanner.Entry `json:"entry"`
	DetectedAt time.Time   `json:"detected_at"`
}

// String returns a human-readable description of the event.
func (e PortEvent) String() string {
	return fmt.Sprintf(
		"[%s] port %s %s at %s",
		e.Type,
		e.Entry.LocalAddress,
		string(e.Type),
		e.DetectedAt.Format(time.RFC3339),
	)
}

// EventsFromDiff converts a scanner.Diff into a slice of PortEvents.
func EventsFromDiff(d scanner.Diff, at time.Time) []PortEvent {
	events := make([]PortEvent, 0, len(d.Added)+len(d.Removed))
	for _, entry := range d.Added {
		events = append(events, PortEvent{
			Type:       EventOpened,
			Entry:      entry,
			DetectedAt: at,
		})
	}
	for _, entry := range d.Removed {
		events = append(events, PortEvent{
			Type:       EventClosed,
			Entry:      entry,
			DetectedAt: at,
		})
	}
	return events
}
