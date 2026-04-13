package alert

import (
	"context"
	"sync"
	"time"
)

// BatchNotifier collects events over a time window and forwards them
// as a single batch to the inner Notifier when the window elapses or
// the buffer reaches maxSize.
type BatchNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	buffer  []Event
	window  time.Duration
	maxSize int
	timer   *time.Timer
}

// NewBatchNotifier creates a BatchNotifier that flushes after window or
// when maxSize events have accumulated.
func NewBatchNotifier(inner Notifier, window time.Duration, maxSize int) *BatchNotifier {
	return &BatchNotifier{
		inner:   inner,
		window:  window,
		maxSize: maxSize,
	}
}

// Send buffers events and flushes when the batch is full or the window fires.
func (b *BatchNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	b.mu.Lock()
	b.buffer = append(b.buffer, events...)
	ready := len(b.buffer) >= b.maxSize

	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, func() {
			_ = b.Flush(ctx)
		})
	}
	b.mu.Unlock()

	if ready {
		return b.Flush(ctx)
	}
	return nil
}

// Flush immediately forwards all buffered events to the inner Notifier.
func (b *BatchNotifier) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.buffer) == 0 {
		b.mu.Unlock()
		return nil
	}
	batch := make([]Event, len(b.buffer))
	copy(batch, b.buffer)
	b.buffer = b.buffer[:0]
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.mu.Unlock()

	return b.inner.Send(ctx, batch)
}
