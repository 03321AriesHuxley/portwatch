package alert

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestQuotaThenBatch verifies that a QuotaNotifier sitting in front of a
// BatchNotifier correctly limits how many events reach the batch.
func TestQuotaThenBatch(t *testing.T) {
	var flushed int32
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&flushed, int32(len(events)))
		return nil
	})

	batch := NewBatchNotifier(sink, 10, 50*time.Millisecond)
	quota := NewQuotaNotifier(batch, 4, time.Minute)

	// Send 6 events; only 4 should reach the batch.
	_ = quota.Send(context.Background(), makeQuotaEvents(6))

	// Flush the batch manually.
	_ = batch.Flush(context.Background())

	if got := atomic.LoadInt32(&flushed); got != 4 {
		t.Fatalf("expected 4 events flushed through batch, got %d", got)
	}
}

// TestThrottleThenQuota verifies that a ThrottleNotifier in front of a
// QuotaNotifier does not bypass the quota on the second allowed send.
func TestThrottleThenQuota(t *testing.T) {
	var received int32
	sink := NotifierFunc(func(_ context.Context, events []Event) error {
		atomic.AddInt32(&received, int32(len(events)))
		return nil
	})

	now := time.Now()
	quota := NewQuotaNotifier(sink, 3, time.Minute)
	quota.now = func() time.Time { return now }

	throttle := NewThrottleNotifier(quota, 10*time.Millisecond)

	// First send: 2 events pass quota and throttle.
	_ = throttle.Send(context.Background(), makeQuotaEvents(2))

	// Advance time past throttle window.
	time.Sleep(20 * time.Millisecond)

	// Second send: 2 events requested but quota only has 1 remaining.
	_ = throttle.Send(context.Background(), makeQuotaEvents(2))

	if got := atomic.LoadInt32(&received); got != 3 {
		t.Fatalf("expected 3 total events (quota=3), got %d", got)
	}
}
