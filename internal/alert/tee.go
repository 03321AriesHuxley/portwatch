package alert

import (
	"context"
	"fmt"
	"strings"
)

// TeeNotifier forwards events to all notifiers, collecting errors from each,
// but unlike FanoutNotifier it always attempts every notifier regardless of
// prior failures and returns a combined error summary.
type TeeNotifier struct {
	notifiers []Notifier
}

// NewTeeNotifier creates a TeeNotifier that sends events to every provided
// notifier. All notifiers are called even if one or more return errors.
func NewTeeNotifier(notifiers ...Notifier) *TeeNotifier {
	return &TeeNotifier{notifiers: notifiers}
}

// Send delivers events to every registered notifier. Errors are accumulated
// and returned as a single combined error. A nil return means all notifiers
// succeeded.
func (t *TeeNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	var errs []string
	for _, n := range t.notifiers {
		if err := n.Send(ctx, events); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("tee: %d notifier(s) failed: %s", len(errs), strings.Join(errs, "; "))
}
