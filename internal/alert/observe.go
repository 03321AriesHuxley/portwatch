package alert

import (
	"context"
	"sync"
	"time"
)

// ObserveEvent records a single observation of a notifier invocation.
type ObserveEvent struct {
	Notifier  string
	SentAt    time.Time
	Duration  time.Duration
	EventCount int
	Err       error
}

// Observer is a callback invoked after each notifier Send call.
type Observer func(ObserveEvent)

// ObserveNotifier wraps a Notifier and calls the provided Observer after
// every Send, recording latency, event count, and any error returned.
// It is useful for instrumenting individual pipeline stages without
// modifying the notifier itself.
type ObserveNotifier struct {
	mu       sync.Mutex
	name     string
	inner    Notifier
	observer Observer
}

// NewObserveNotifier returns an ObserveNotifier that wraps inner and calls
// obs after every Send. name is used to identify the notifier in the
// ObserveEvent so callers can distinguish multiple wrapped notifiers.
func NewObserveNotifier(name string, inner Notifier, obs Observer) *ObserveNotifier {
	if obs == nil {
		obs = func(ObserveEvent) {}
	}
	return &ObserveNotifier{
		name:     name,
		inner:    inner,
		observer: obs,
	}
}

// Send delegates to the wrapped Notifier and then calls the Observer with
// timing and result information. The observer is always called, even when
// the inner notifier returns an error or the context is cancelled.
func (o *ObserveNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	err := o.inner.Send(ctx, events)
	dur := time.Since(start)

	o.mu.Lock()
	obs := o.observer
	o.mu.Unlock()

	obs(ObserveEvent{
		Notifier:   o.name,
		SentAt:     start,
		Duration:   dur,
		EventCount: len(events),
		Err:        err,
	})

	return err
}

// SetObserver replaces the observer function at runtime. It is safe for
// concurrent use.
func (o *ObserveNotifier) SetObserver(obs Observer) {
	if obs == nil {
		obs = func(ObserveEvent) {}
	}
	o.mu.Lock()
	o.observer = obs
	o.mu.Unlock()
}
