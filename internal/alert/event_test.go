package alert_test

import (
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeEntry(addr string) scanner.Entry {
	return scanner.Entry{LocalAddress: addr, State: "0A"}
}

func TestEventsFromDiff_Empty(t *testing.T) {
	d := scanner.Diff{}
	events := alert.EventsFromDiff(d, time.Now())
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestEventsFromDiff_Added(t *testing.T) {
	d := scanner.Diff{
		Added: []scanner.Entry{makeEntry("0.0.0.0:8080")},
	}
	at := time.Now()
	events := alert.EventsFromDiff(d, at)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != alert.EventOpened {
		t.Errorf("expected EventOpened, got %s", events[0].Type)
	}
	if events[0].DetectedAt != at {
		t.Errorf("unexpected DetectedAt")
	}
}

func TestEventsFromDiff_Removed(t *testing.T) {
	d := scanner.Diff{
		Removed: []scanner.Entry{makeEntry("0.0.0.0:9090")},
	}
	events := alert.EventsFromDiff(d, time.Now())
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != alert.EventClosed {
		t.Errorf("expected EventClosed, got %s", events[0].Type)
	}
}

func TestEventsFromDiff_Mixed(t *testing.T) {
	d := scanner.Diff{
		Added:   []scanner.Entry{makeEntry("0.0.0.0:80"), makeEntry("0.0.0.0:443")},
		Removed: []scanner.Entry{makeEntry("0.0.0.0:22")},
	}
	events := alert.EventsFromDiff(d, time.Now())
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestPortEvent_String(t *testing.T) {
	at := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	ev := alert.PortEvent{
		Type:       alert.EventOpened,
		Entry:      makeEntry("0.0.0.0:8080"),
		DetectedAt: at,
	}
	s := ev.String()
	if !strings.Contains(s, "opened") {
		t.Errorf("expected 'opened' in string, got: %s", s)
	}
	if !strings.Contains(s, "8080") {
		t.Errorf("expected port in string, got: %s", s)
	}
}
