package alert

import (
	"context"
	"fmt"
	"time"
)

// WindowNotifier wraps a Notifier and suppresses sends when the number of
// events dispatched within the sliding window exceeds a configured threshold.
// Once the count drops back below the threshold, sends resume normally.
type WindowNotifier struct {
	inner     Notifier
	counter   *WindowCounter
	threshold int
}

// NewWindowNotifier creates a WindowNotifier that allows at most threshold
// events within the given window duration before suppressing further sends.
func NewWindowNotifier(inner Notifier, window time.Duration, buckets, threshold int) *WindowNotifier {
	return &WindowNotifier{
		inner:     inner,
		counter:   NewWindowCounter(window, buckets),
		threshold: threshold,
	}
}

// Send forwards events to the inner Notifier only when the sliding-window
// event count is within the configured threshold. Events that would exceed the
// threshold are dropped and an error is returned so callers can log or handle
// the suppression.
func (w *WindowNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return w.inner.Send(ctx, events)
	}

	current := w.counter.Count()
	if current+len(events) > w.threshold {
		return fmt.Errorf(
			"window notifier: suppressing %d event(s); window count %d exceeds threshold %d",
			len(events), current, w.threshold,
		)
	}

	w.counter.Add(len(events))
	return w.inner.Send(ctx, events)
}
