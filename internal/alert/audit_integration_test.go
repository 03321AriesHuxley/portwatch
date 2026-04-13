package alert_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/yourusername/portwatch/internal/alert"
	"github.com/yourusername/portwatch/internal/scanner"
)

// TestAuditLog_WithRetryIntegration verifies that AuditLog correctly wraps a
// RetryNotifier and records each top-level Send attempt (not internal retries).
func TestAuditLog_WithRetryIntegration(t *testing.T) {
	var buf bytes.Buffer

	// inner succeeds on first attempt
	cap := &captureNotifier{}
	retrier := alert.NewRetryNotifier(cap, 3)
	al := alert.NewAuditLog(retrier, "retry-wrapped", &buf)

	events := []alert.Event{
		{
			Kind: alert.EventAdded,
			Entry: scanner.Entry{
				LocalAddress: "127.0.0.1",
				LocalPort:    8080,
				Protocol:     "tcp",
			},
		},
	}

	if err := al.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	recs := al.Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(recs))
	}
	if recs[0].Err != nil {
		t.Errorf("expected success record, got err: %v", recs[0].Err)
	}
	if recs[0].EventCount != 1 {
		t.Errorf("expected 1 event in record, got %d", recs[0].EventCount)
	}

	output := buf.String()
	if !strings.Contains(output, "retry-wrapped") {
		t.Errorf("expected notifier name in audit output, got: %s", output)
	}
	if !strings.Contains(output, "ok") {
		t.Errorf("expected 'ok' status in audit output, got: %s", output)
	}
}
