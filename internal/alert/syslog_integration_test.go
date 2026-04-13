//go:build integration

package alert_test

import (
	"context"
	"log/syslog"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

// TestSyslogNotifier_Integration sends a burst of events through the full
// notifier stack (throttle → syslog) and verifies no errors occur.
// Run with: go test -tags integration ./internal/alert/...
func TestSyslogNotifier_Integration(t *testing.T) {
	sn, err := alert.NewSyslogNotifier(
		alert.WithSyslogTag("portwatch-integration"),
		alert.WithSyslogPriority(syslog.LOG_DEBUG|syslog.LOG_LOCAL0),
	)
	if err != nil {
		t.Skipf("syslog unavailable: %v", err)
	}
	t.Cleanup(func() { _ = sn.Close() })

	throttled := alert.NewThrottleNotifier(sn, 100*time.Millisecond)

	events := []alert.Event{
		{Kind: alert.EventAdded, Entry: scanner.Entry{Protocol: "tcp", LocalAddr: "0.0.0.0", LocalPort: 2222}},
		{Kind: alert.EventAdded, Entry: scanner.Entry{Protocol: "udp", LocalAddr: "0.0.0.0", LocalPort: 5353}},
	}

	ctx := context.Background()

	// First send should pass through.
	if err := throttled.Send(ctx, events); err != nil {
		t.Fatalf("first send: %v", err)
	}

	// Second send within window should be suppressed (no error).
	if err := throttled.Send(ctx, events); err != nil {
		t.Fatalf("throttled send: %v", err)
	}

	// After window, send should pass through again.
	time.Sleep(150 * time.Millisecond)
	if err := throttled.Send(ctx, events); err != nil {
		t.Fatalf("post-window send: %v", err)
	}
}
