package alert

import (
	"context"
	"time"
)

// ScheduleWindow defines a time-of-day window during which notifications are allowed.
type ScheduleWindow struct {
	// Start is the hour (0-23) at which the window opens.
	Start int
	// End is the hour (0-23) at which the window closes (exclusive).
	End int
}

// Active reports whether the given time falls within the window.
// If Start == End the window is considered always-open.
func (w ScheduleWindow) Active(t time.Time) bool {
	if w.Start == w.End {
		return true
	}
	h := t.Hour()
	if w.Start < w.End {
		return h >= w.Start && h < w.End
	}
	// Wraps midnight, e.g. 22-06
	return h >= w.Start || h < w.End
}

// scheduleNotifier wraps a Notifier and only forwards events when the current
// time falls within one of the configured ScheduleWindows. Events that arrive
// outside every window are silently dropped.
type scheduleNotifier struct {
	inner   Notifier
	windows []ScheduleWindow
	now     func() time.Time
}

// ScheduleOption configures a scheduleNotifier.
type ScheduleOption func(*scheduleNotifier)

// WithScheduleClock overrides the clock used for window evaluation. Useful in
// tests.
func WithScheduleClock(fn func() time.Time) ScheduleOption {
	return func(s *scheduleNotifier) { s.now = fn }
}

// NewScheduleNotifier returns a Notifier that forwards events to inner only
// when the current wall-clock hour falls within at least one of the provided
// windows. If no windows are provided all events are forwarded.
func NewScheduleNotifier(inner Notifier, windows []ScheduleWindow, opts ...ScheduleOption) Notifier {
	sn := &scheduleNotifier{
		inner:   inner,
		windows: windows,
		now:     time.Now,
	}
	for _, o := range opts {
		o(sn)
	}
	return sn
}

// Send forwards events to the inner Notifier when the current time is within
// an active window. If no windows are configured, or if the current time
// matches at least one window, the full event slice is forwarded unchanged.
func (s *scheduleNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if len(s.windows) == 0 {
		return s.inner.Send(ctx, events)
	}
	now := s.now()
	for _, w := range s.windows {
		if w.Active(now) {
			return s.inner.Send(ctx, events)
		}
	}
	// Outside all windows — drop silently.
	return nil
}
