package alert

import (
	"context"
	"math/rand"
	"time"
)

// JitterNotifier wraps a Notifier and introduces a random delay before each
// Send call. This helps spread bursts across distributed instances.
type JitterNotifier struct {
	inner  Notifier
	minDelay time.Duration
	maxDelay time.Duration
	rng    *rand.Rand
}

// NewJitterNotifier returns a Notifier that waits a random duration in
// [minDelay, maxDelay) before delegating to inner.
func NewJitterNotifier(inner Notifier, minDelay, maxDelay time.Duration) Notifier {
	if maxDelay <= minDelay {
		maxDelay = minDelay + time.Millisecond
	}
	return &JitterNotifier{
		inner:    inner,
		minDelay: minDelay,
		maxDelay: maxDelay,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (j *JitterNotifier) Send(ctx context.Context, events []Event) error {
	window := int64(j.maxDelay - j.minDelay)
	jitter := time.Duration(j.rng.Int63n(window)) + j.minDelay
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(jitter):
	}
	return j.inner.Send(ctx, events)
}
