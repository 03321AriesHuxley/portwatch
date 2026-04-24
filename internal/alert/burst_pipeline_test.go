package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// TestBurst_InPipeline verifies BurstNotifier integrates correctly as a
// pipeline stage, forwarding burst events and dropping excess.
func TestBurst_InPipeline(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	now := time.Now()
	clock := func() time.Time { return now }

	p := alert.NewPipeline(
		alert.NewBurstNotifier(sink, 2, 10*time.Second, 30*time.Second,
			alert.WithBurstClock(clock)),
	)

	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: scanner.Entry{Port: 80, Protocol: "tcp", Address: "0.0.0.0"}},
		{Kind: alert.EventAdded, Entry: scanner.Entry{Port: 443, Protocol: "tcp", Address: "0.0.0.0"}},
		{Kind: alert.EventAdded, Entry: scanner.Entry{Port: 8080, Protocol: "tcp", Address: "0.0.0.0"}},
	}

	if err := p.Send(context.Background(), events); err != nil {
		t.Fatalf("pipeline send error: %v", err)
	}

	if len(received) != 2 {
		t.Errorf("expected 2 events through burst limit, got %d", len(received))
	}
}

// TestBurst_WithThrottle verifies BurstNotifier followed by ThrottleNotifier
// correctly limits the total event throughput.
func TestBurst_WithThrottle(t *testing.T) {
	var received []alert.Event
	sink := alert.NotifierFunc(func(_ context.Context, evs []alert.Event) error {
		received = append(received, evs...)
		return nil
	})

	now := time.Now()
	clock := func() time.Time { return now }

	burst := alert.NewBurstNotifier(
		alert.NewThrottleNotifier(sink, 1*time.Second, alert.WithThrottleClock(clock)),
		3, 200*time.Millisecond, 10*time.Second,
		alert.WithBurstClock(clock),
	)

	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: scanner.Entry{Port: 22, Protocol: "tcp", Address: "0.0.0.0"}},
		{Kind: alert.EventAdded, Entry: scanner.Entry{Port: 80, Protocol: "tcp", Address: "0.0.0.0"}},
	}

	if err := burst.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both events pass burst; throttle allows first batch through.
	if len(received) == 0 {
		t.Error("expected at least one event to reach sink")
	}
}
