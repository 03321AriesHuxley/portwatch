package alert

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// FanoutNotifier sends events to all registered notifiers concurrently,
// collecting any errors and returning a combined error if any notifier fails.
type FanoutNotifier struct {
	notifiers []Notifier
}

// NewFanoutNotifier returns a FanoutNotifier that dispatches to all given notifiers in parallel.
func NewFanoutNotifier(notifiers ...Notifier) *FanoutNotifier {
	return &FanoutNotifier{notifiers: notifiers}
}

// Send dispatches events to all notifiers concurrently and waits for all to complete.
// If one or more notifiers return an error, a combined error is returned.
func (f *FanoutNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if len(f.notifiers) == 0 {
		return nil
	}

	type result struct {
		idx int
		err error
	}

	results := make(chan result, len(f.notifiers))
	var wg sync.WaitGroup

	for i, n := range f.notifiers {
		wg.Add(1)
		go func(idx int, notifier Notifier) {
			defer wg.Done()
			err := notifier.Send(ctx, events)
			results <- result{idx: idx, err: err}
		}(i, n)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []string
	for r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("notifier[%d]: %v", r.idx, r.err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("fanout errors: %s", strings.Join(errs, "; "))
	}
	return nil
}
