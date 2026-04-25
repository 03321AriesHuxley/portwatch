package alert

import (
	"context"
	"time"
)

// StaggerNotifier forwards events to the inner notifier but introduces a
// fixed delay between each individual event, spreading load over time.
type StaggerNotifier struct {
	inner    Notifier
	delay    time.Duration
	sleepFn  func(context.Context, time.Duration) error
}

// NewStaggerNotifier creates a StaggerNotifier that waits delay between
// each event before forwarding to inner. A zero or negative delay is
// clamped to 1ms to avoid a busy loop.
func NewStaggerNotifier(inner Notifier, delay time.Duration) *StaggerNotifier {
	if delay <= 0 {
		delay = time.Millisecond
	}
	return &StaggerNotifier{
		inner: inner,
		delay: delay,
		sleepFn: func(ctx context.Context, d time.Duration) error {
			select {
			case <-time.After(d):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
}

// Send forwards each event individually to the inner notifier, sleeping
// delay between events. If the context is cancelled the send stops early
// and returns the context error.
func (s *StaggerNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	for i, ev := range events {
		if err := s.inner.Send(ctx, []Event{ev}); err != nil {
			return err
		}
		if i < len(events)-1 {
			if err := s.sleepFn(ctx, s.delay); err != nil {
				return err
			}
		}
	}
	return nil
}
