package alert

import (
	"context"
	"fmt"
	"time"
)

// SummaryNotifier wraps a Notifier and periodically emits a summary of
// all events seen since the last flush, instead of forwarding them
// individually. It is useful for reducing noise on high-churn hosts.
type SummaryNotifier struct {
	inner    Notifier
	window   time.Duration
	buf      []Event
	lastSent time.Time
}

// NewSummaryNotifier returns a SummaryNotifier that batches events and
// forwards a single summary message to inner every window duration.
func NewSummaryNotifier(inner Notifier, window time.Duration) *SummaryNotifier {
	return &SummaryNotifier{
		inner:  inner,
		window: window,
	}
}

// Send buffers the provided events. When the configured window has
// elapsed since the last flush the buffer is drained and a single
// summarised event slice is forwarded to the inner Notifier.
func (s *SummaryNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	s.buf = append(s.buf, events...)

	if time.Since(s.lastSent) < s.window {
		return nil
	}

	if err := s.flush(ctx); err != nil {
		return fmt.Errorf("summary flush: %w", err)
	}

	return nil
}

// Flush forces an immediate drain of the buffer regardless of the
// window, forwarding any buffered events to the inner Notifier.
func (s *SummaryNotifier) Flush(ctx context.Context) error {
	if len(s.buf) == 0 {
		return nil
	}
	return s.flush(ctx)
}

func (s *SummaryNotifier) flush(ctx context.Context) error {
	payload := make([]Event, len(s.buf))
	copy(payload, s.buf)
	s.buf = s.buf[:0]
	s.lastSent = time.Now()
	return s.inner.Send(ctx, payload)
}
