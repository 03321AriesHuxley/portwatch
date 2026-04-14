package alert

import (
	"testing"
	"time"
)

func TestWindowCounter_StartsAtZero(t *testing.T) {
	w := NewWindowCounter(time.Second, 10)
	if got := w.Count(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestWindowCounter_AddsAndCounts(t *testing.T) {
	w := NewWindowCounter(time.Second, 10)
	w.Add(3)
	w.Add(2)
	if got := w.Count(); got != 5 {
		t.Fatalf("expected 5, got %d", got)
	}
}

func TestWindowCounter_Reset(t *testing.T) {
	w := NewWindowCounter(time.Second, 10)
	w.Add(7)
	w.Reset()
	if got := w.Count(); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestWindowCounter_ExpiresOldBuckets(t *testing.T) {
	// Use a very short window so we can simulate expiry.
	// 4 buckets over 40ms => each bucket is 10ms.
	w := NewWindowCounter(40*time.Millisecond, 4)
	w.Add(5)

	// Wait longer than the full window so all buckets rotate out.
	time.Sleep(50 * time.Millisecond)

	if got := w.Count(); got != 0 {
		t.Fatalf("expected 0 after window expiry, got %d", got)
	}
}

func TestWindowCounter_PartialExpiry(t *testing.T) {
	// 4 buckets over 80ms => each bucket is 20ms.
	w := NewWindowCounter(80*time.Millisecond, 4)
	w.Add(10)

	// Sleep for just over one bucket interval; the first bucket should
	// have rotated but the count should still be present in older slots.
	time.Sleep(25 * time.Millisecond)
	w.Add(3)

	got := w.Count()
	if got != 13 {
		t.Fatalf("expected 13, got %d", got)
	}
}

func TestWindowCounter_ConcurrentSafe(t *testing.T) {
	w := NewWindowCounter(time.Second, 10)
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			w.Add(1)
			_ = w.Count()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	if got := w.Count(); got != 50 {
		t.Fatalf("expected 50 concurrent adds, got %d", got)
	}
}

func TestNewWindowCounter_InvalidBuckets(t *testing.T) {
	// Should not panic with zero buckets.
	w := NewWindowCounter(time.Second, 0)
	w.Add(1)
	if got := w.Count(); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}
