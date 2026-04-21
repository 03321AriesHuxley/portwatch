package alert

import (
	"context"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

// TestRouter_Integration verifies that a realistic routing topology works
// end-to-end: critical ports go to a dedicated notifier, everything else
// goes to a catch-all, and a fallback handles unmatched events.
func TestRouter_Integration(t *testing.T) {
	var critical, general, fallback []Event

	criticalN := NotifierFunc(func(_ context.Context, evs []Event) error {
		critical = append(critical, evs...)
		return nil
	})
	generalN := NotifierFunc(func(_ context.Context, evs []Event) error {
		general = append(general, evs...)
		return nil
	})
	fallbackN := NotifierFunc(func(_ context.Context, evs []Event) error {
		fallback = append(fallback, evs...)
		return nil
	})

	routes := []route{
		{predicate: HasPort(22), notifier: criticalN},
		{predicate: HasPort(443), notifier: criticalN},
		{predicate: AnyEvent, notifier: generalN},
	}
	r := NewRouter(routes, WithFallback(fallbackN))

	events := []Event{
		{Kind: "added", Entry: scanner.Entry{Port: 22, Addr: "0.0.0.0", Proto: scanner.TCP}},
		{Kind: "added", Entry: scanner.Entry{Port: 8080, Addr: "0.0.0.0", Proto: scanner.TCP}},
	}

	if err := r.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Port 22 matched criticalN and AnyEvent (generalN), so both should fire.
	if len(critical) != 2 {
		t.Errorf("critical: expected 2 events, got %d", len(critical))
	}
	if len(general) != 2 {
		t.Errorf("general: expected 2 events, got %d", len(general))
	}
	// All routes matched, so fallback should NOT be called.
	if len(fallback) != 0 {
		t.Errorf("fallback: expected 0 events, got %d", len(fallback))
	}
}
