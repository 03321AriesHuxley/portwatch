package alert

import (
	"fmt"
	"strings"
	"time"
)

// FormatEvent returns a human-readable string representation of an Event.
func FormatEvent(e Event) string {
	verb := "bound"
	if e.Kind == EventRemoved {
		verb = "released"
	}
	return fmt.Sprintf("[%s] port %s/%d %s on %s",
		e.Time.UTC().Format(time.RFC3339),
		strings.ToUpper(e.Entry.Protocol),
		e.Entry.Port,
		verb,
		e.Entry.Address,
	)
}

// FormatEvents returns a multi-line string with one formatted event per line.
func FormatEvents(events []Event) string {
	if len(events) == 0 {
		return ""
	}
	lines := make([]string, 0, len(events))
	for _, e := range events {
		lines = append(lines, FormatEvent(e))
	}
	return strings.Join(lines, "\n")
}

// FormatSummary returns a compact summary line describing a batch of events.
func FormatSummary(events []Event) string {
	if len(events) == 0 {
		return "no port changes detected"
	}
	added, removed := 0, 0
	for _, e := range events {
		switch e.Kind {
		case EventAdded:
			added++
		case EventRemoved:
			removed++
		}
	}
	parts := []string{}
	if added > 0 {
		parts = append(parts, fmt.Sprintf("%d added", added))
	}
	if removed > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", removed))
	}
	return fmt.Sprintf("port changes: %s", strings.Join(parts, ", "))
}
