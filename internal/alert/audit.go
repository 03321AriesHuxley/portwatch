package alert

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditRecord captures a single notification attempt and its outcome.
type AuditRecord struct {
	Timestamp time.Time
	Notifier  string
	EventCount int
	Err       error
}

// AuditLog records notification outcomes to a writer for operational tracing.
type AuditLog struct {
	mu      sync.Mutex
	w       io.Writer
	inner   Notifier
	name    string
	records []AuditRecord
}

// NewAuditLog wraps a Notifier and writes audit records to w.
// name identifies the wrapped notifier in log output.
func NewAuditLog(inner Notifier, name string, w io.Writer) *AuditLog {
	if w == nil {
		w = os.Stderr
	}
	return &AuditLog{inner: inner, name: name, w: w}
}

// Send forwards events to the inner notifier and records the result.
func (a *AuditLog) Send(ctx context.Context, events []Event) error {
	start := time.Now()
	err := a.inner.Send(ctx, events)
	rec := AuditRecord{
		Timestamp:  start,
		Notifier:   a.name,
		EventCount: len(events),
		Err:        err,
	}
	a.mu.Lock()
	a.records = append(a.records, rec)
	a.mu.Unlock()

	status := "ok"
	if err != nil {
		status = fmt.Sprintf("error: %v", err)
	}
	fmt.Fprintf(a.w, "[audit] %s notifier=%s events=%d status=%s\n",
		start.Format(time.RFC3339), a.name, len(events), status)
	return err
}

// Records returns a snapshot of all recorded audit entries.
func (a *AuditLog) Records() []AuditRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditRecord, len(a.records))
	copy(out, a.records)
	return out
}
