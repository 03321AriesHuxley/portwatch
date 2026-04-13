package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

// TestThrottleThenBatch verifies that a ThrottleNotifier wrapping a
// BatchNotifier correctly suppresses rapid sends and batches survivors.
func TestThrottleThenBatch(t *testing.T) {
	inner := &capturingNotifier{}
	batch := alert.NewBatchNotifier(inner, 5*time.Second, 10)
	throttle := alert.NewThrottleNotifier(batch, 50*time.Millisecond)
	ctx := context.Background()

	evt := []alert.Event{{Kind: alert.EventAdded}}

	// First send passes throttle → buffered in batch
	_ = throttle.Send(ctx, evt)
	// Second send within window → dropped by throttle
	_ = throttle.Send(ctx, evt)

	// Manually flush batch to verify only one event was buffered
	_ = batch.Flush(ctx)

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 batch flush, got %d", inner.calls.Load())
	}
	if len(inner.events) != 1 {
		t.Fatalf("expected 1 event in batch, got %d", len(inner.events))
	}
}

// TestBatchThenThrottle verifies that a BatchNotifier wrapping a
// ThrottleNotifier accumulates then throttles the forwarded batch.
func TestBatchThenThrottle(t *testing.T) {
	inner := &capturingNotifier{}
	throttle := alert.NewThrottleNotifier(inner, 10*time.Second)
	batch := alert.NewBatchNotifier(throttle, 5*time.Second, 2)
	ctx := context.Background()

	evt := []alert.Event{{Kind: alert.EventAdded}}

	// Fill batch to maxSize — triggers flush to throttle
	_ = batch.Send(ctx, evt)
	_ = batch.Send(ctx, evt)

	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls.Load())
	}

	// Second batch flush should be throttled
	_ = batch.Send(ctx, evt)
	_ = batch.Send(ctx, evt)

	if inner.calls.Load() != 1 {
		t.Fatalf("expected still 1 inner call after throttle, got %d", inner.calls.Load())
	}
}
