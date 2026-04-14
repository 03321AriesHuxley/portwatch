package alert

import (
	"sync"
	"time"
)

// WindowCounter tracks event counts within a sliding time window.
// It is safe for concurrent use.
type WindowCounter struct {
	mu       sync.Mutex
	buckets  []int
	size     int
	interval time.Duration
	last     time.Time
}

// NewWindowCounter creates a WindowCounter that divides the given duration
// into the specified number of buckets for sliding-window counting.
func NewWindowCounter(window time.Duration, buckets int) *WindowCounter {
	if buckets < 1 {
		buckets = 1
	}
	return &WindowCounter{
		buckets:  make([]int, buckets),
		size:     buckets,
		interval: window / time.Duration(buckets),
		last:     time.Now(),
	}
}

// Add records n occurrences in the current bucket.
func (w *WindowCounter) Add(n int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.advance(time.Now())
	w.buckets[0] += n
}

// Count returns the total number of occurrences within the sliding window.
func (w *WindowCounter) Count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.advance(time.Now())
	total := 0
	for _, v := range w.buckets {
		total += v
	}
	return total
}

// Reset clears all buckets.
func (w *WindowCounter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets = make([]int, w.size)
	w.last = time.Now()
}

// advance rotates buckets forward based on elapsed time.
func (w *WindowCounter) advance(now time.Time) {
	elapsed := now.Sub(w.last)
	steps := int(elapsed / w.interval)
	if steps <= 0 {
		return
	}
	if steps >= w.size {
		w.buckets = make([]int, w.size)
	} else {
		rotated := make([]int, w.size)
		copy(rotated[steps:], w.buckets[:w.size-steps])
		w.buckets = rotated
	}
	w.last = w.last.Add(time.Duration(steps) * w.interval)
}
