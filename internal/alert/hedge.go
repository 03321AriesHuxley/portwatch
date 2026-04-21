package alert

import (
	"context"
	"sync"
	"time"
)

// HedgeNotifier sends to a primary notifier and, if it does not complete within
// a hedge delay, simultaneously fires a secondary notifier. The first
// successful response wins; the other branch is cancelled.
type HedgeNotifier struct {
	primary   Notifier
	secondary Notifier
	delay     time.Duration
}

// NewHedgeNotifier returns a HedgeNotifier that will launch a secondary
// notification attempt after delay if the primary has not yet returned.
func NewHedgeNotifier(primary, secondary Notifier, delay time.Duration) *HedgeNotifier {
	return &HedgeNotifier{
		primary:   primary,
		secondary: secondary,
		delay:     delay,
	}
}

type hedgeResult struct {
	err error
}

// Send dispatches events to the primary notifier. If the primary does not
// complete within the configured delay the secondary is also invoked. The
// first nil-error result is returned; if both fail the primary error is used.
func (h *HedgeNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan hedgeResult, 2)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := h.primary.Send(ctx, events)
		results <- hedgeResult{err: err}
	}()

	timer := time.NewTimer(h.delay)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := h.secondary.Send(ctx, events)
				results <- hedgeResult{err: err}
			}()
		case <-ctx.Done():
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	got := 0
	for r := range results {
		got++
		if r.err == nil {
			cancel()
			return nil
		}
		if firstErr == nil {
			firstErr = r.err
		}
		if got == 2 {
			break
		}
	}
	return firstErr
}
