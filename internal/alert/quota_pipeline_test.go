package alert

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestPipeline_WithQuota ensures the PipelineBuilder can incorporate a
// QuotaNotifier stage and that the quota is enforced end-to-end.
func TestPipeline_WithQuota(t *testing.T) {
	var count int32
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&count, int32(len(events)))
		return nil
	})

	quota := NewQuotaNotifier(sink, 5, time.Minute)

	pipeline := NewPipelineBuilder().
		Add(quota).
		Build()

	events := makeQuotaEvents(8)

	// First call: 5 pass, 3 are truncated.
	_ = pipeline.Send(context.Background(), events)

	if got := atomic.LoadInt32(&count); got != 5 {
		t.Fatalf("expected 5 events through pipeline, got %d", got)
	}

	if rem := quota.Remaining(); rem != 0 {
		t.Fatalf("expected 0 remaining after quota exhausted, got %d", rem)
	}
}

// TestPipeline_QuotaResetIntegration confirms that after the quota window
// expires the pipeline forwards events again.
func TestPipeline_QuotaResetIntegration(t *testing.T) {
	var count int32
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&count, int32(len(events)))
		return nil
	})

	now := time.Now()
	quota := NewQuotaNotifier(sink, 2, 30*time.Millisecond)
	quota.now = func() time.Time { return now }

	pipeline := NewPipelineBuilder().Add(quota).Build()

	_ = pipeline.Send(context.Background(), makeQuotaEvents(2))

	// Advance clock past window.
	quota.now = func() time.Time { return now.Add(60 * time.Millisecond) }

	_ = pipeline.Send(context.Background(), makeQuotaEvents(2))

	if got := atomic.LoadInt32(&count); got != 4 {
		t.Fatalf("expected 4 total events after window reset, got %d", got)
	}
}
