package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/scanner"
)

func makeFormatterEntry(proto, addr string, port uint16) scanner.Entry {
	return scanner.Entry{Protocol: proto, Address: addr, Port: port}
}

var fixedTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func TestFormatEvent_Added(t *testing.T) {
	e := Event{
		Kind:  EventAdded,
		Entry: makeFormatterEntry("tcp", "0.0.0.0", 8080),
		Time:  fixedTime,
	}
	out := FormatEvent(e)
	if !strings.Contains(out, "TCP") {
		t.Errorf("expected protocol TCP in output, got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	if !strings.Contains(out, "bound") {
		t.Errorf("expected verb 'bound' in output, got: %s", out)
	}
}

func TestFormatEvent_Removed(t *testing.T) {
	e := Event{
		Kind:  EventRemoved,
		Entry: makeFormatterEntry("udp", "127.0.0.1", 53),
		Time:  fixedTime,
	}
	out := FormatEvent(e)
	if !strings.Contains(out, "released") {
		t.Errorf("expected verb 'released' in output, got: %s", out)
	}
	if !strings.Contains(out, "53") {
		t.Errorf("expected port 53 in output, got: %s", out)
	}
}

func TestFormatEvents_Empty(t *testing.T) {
	out := FormatEvents(nil)
	if out != "" {
		t.Errorf("expected empty string for no events, got: %q", out)
	}
}

func TestFormatEvents_MultiLine(t *testing.T) {
	events := []Event{
		{Kind: EventAdded, Entry: makeFormatterEntry("tcp", "0.0.0.0", 80), Time: fixedTime},
		{Kind: EventRemoved, Entry: makeFormatterEntry("tcp", "0.0.0.0", 443), Time: fixedTime},
	}
	out := FormatEvents(events)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %s", len(lines), out)
	}
}

func TestFormatSummary_Empty(t *testing.T) {
	out := FormatSummary(nil)
	if out != "no port changes detected" {
		t.Errorf("unexpected summary for empty events: %q", out)
	}
}

func TestFormatSummary_Mixed(t *testing.T) {
	events := []Event{
		{Kind: EventAdded, Entry: makeFormatterEntry("tcp", "0.0.0.0", 80), Time: fixedTime},
		{Kind: EventAdded, Entry: makeFormatterEntry("tcp", "0.0.0.0", 8080), Time: fixedTime},
		{Kind: EventRemoved, Entry: makeFormatterEntry("udp", "0.0.0.0", 53), Time: fixedTime},
	}
	out := FormatSummary(events)
	if !strings.Contains(out, "2 added") {
		t.Errorf("expected '2 added' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 removed") {
		t.Errorf("expected '1 removed' in summary, got: %s", out)
	}
}
