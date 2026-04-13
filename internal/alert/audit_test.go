package alert_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/portwatch/internal/alert"
	"github.com/yourusername/portwatch/internal/scanner"
)

func makeAuditEvent() alert.Event {
	return alert.Event{
		Kind: alert.EventAdded,
		Entry: scanner.Entry{
			LocalAddress: "0.0.0.0",
			LocalPort:    9090,
			Protocol:     "tcp",
		},
	}
}

func TestAuditLog_RecordsSuccess(t *testing.T) {
	var buf bytes.Buffer
	inner := &captureNotifier{}
	al := alert.NewAuditLog(inner, "test-notifier", &buf)

	events := []alert.Event{makeAuditEvent()}
	if err := al.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	recs := al.Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if recs[0].Err != nil {
		t.Errorf("expected nil error, got %v", recs[0].Err)
	}
	if recs[0].EventCount != 1 {
		t.Errorf("expected event count 1, got %d", recs[0].EventCount)
	}
	if recs[0].Notifier != "test-notifier" {
		t.Errorf("unexpected notifier name: %s", recs[0].Notifier)
	}
	if buf.Len() == 0 {
		t.Error("expected audit output in writer")
	}
}

func TestAuditLog_RecordsFailure(t *testing.T) {
	var buf bytes.Buffer
	sentinel := errors.New("send failed")
	inner := &failNotifier{err: sentinel}
	al := alert.NewAuditLog(inner, "failing", &buf)

	err := al.Send(context.Background(), []alert.Event{makeAuditEvent()})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}

	recs := al.Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if !errors.Is(recs[0].Err, sentinel) {
		t.Errorf("expected sentinel in record, got %v", recs[0].Err)
	}
}

func TestAuditLog_TimestampIsRecent(t *testing.T) {
	before := time.Now()
	var buf bytes.Buffer
	al := alert.NewAuditLog(&captureNotifier{}, "ts-test", &buf)
	_ = al.Send(context.Background(), nil)
	after := time.Now()

	recs := al.Records()
	if len(recs) == 0 {
		t.Fatal("no records")
	}
	ts := recs[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestAuditLog_NilWriterDefaultsToStderr(t *testing.T) {
	al := alert.NewAuditLog(&captureNotifier{}, "nil-writer", nil)
	if al == nil {
		t.Fatal("expected non-nil AuditLog")
	}
	// Should not panic when sending
	_ = al.Send(context.Background(), nil)
}
