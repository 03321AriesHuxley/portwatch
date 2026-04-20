package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QuotaNotifier enforces a maximum number of notifications per rolling time
// window. Once the quota is exhausted, events are dropped until the window
// resets. Unlike ThrottleNotifier (which gates on time between sends),
// QuotaNotifier counts individual events across all sends in the window.
type QuotaNotifier struct {
	inner    Notifier
	max      int
	window   time.Duration
	mu       sync.Mutex
	count    int
	windowAt time.Time
	now      func() time.Time
}

// NewQuotaNotifier wraps inner and allows at most max events per window.
func NewQuotaNotifier(inner Notifier, max int, window time.Duration) *QuotaNotifier {
	return &QuotaNotifier{
		inner:  inner,
		max:    max,
		window: window,
		now:    time.Now,
	}
}

func (q *QuotaNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	q.mu.Lock()
	now := q.now()
	if now.After(q.windowAt.Add(q.window)) {
		q.count = 0
		q.windowAt = now
	}
	remaining := q.max - q.count
	q.mu.Unlock()

	if remaining <= 0 {
		return fmt.Errorf("quota exhausted: %d/%d events used in window", q.count, q.max)
	}

	allowed := events
	if len(events) > remaining {
		allowed = events[:remaining]
	}

	q.mu.Lock()
	q.count += len(allowed)
	q.mu.Unlock()

	return q.inner.Send(ctx, allowed)
}

// Remaining returns the number of events still allowed in the current window.
func (q *QuotaNotifier) Remaining() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := q.now()
	if now.After(q.windowAt.Add(q.window)) {
		return q.max
	}
	r := q.max - q.count
	if r < 0 {
		return 0
	}
	return r
}
