package alert

import (
	"context"
	"sync"
	"time"
)

// SuppressNotifier wraps a Notifier and suppresses all forwarding during
// a scheduled time window (e.g. maintenance windows). Outside the window
// events are forwarded normally.
type SuppressNotifier struct {
	inner    Notifier
	mu       sync.RWMutex
	windows  []suppressWindow
	now      func() time.Time
}

type suppressWindow struct {
	start time.Time
	end   time.Time
}

// NewSuppressNotifier creates a SuppressNotifier wrapping inner.
func NewSuppressNotifier(inner Notifier) *SuppressNotifier {
	return &SuppressNotifier{
		inner: inner,
		now:   time.Now,
	}
}

// AddWindow registers a suppression window. Events arriving between start
// and end (inclusive) are silently dropped.
func (s *SuppressNotifier) AddWindow(start, end time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows = append(s.windows, suppressWindow{start: start, end: end})
}

// ClearWindows removes all registered suppression windows.
func (s *SuppressNotifier) ClearWindows() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows = s.windows[:0]
}

// IsSuppressed reports whether the current time falls within any registered window.
func (s *SuppressNotifier) IsSuppressed() bool {
	now := s.now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.windows {
		if !now.Before(w.start) && !now.After(w.end) {
			return true
		}
	}
	return false
}

// Send forwards events to the inner Notifier unless currently suppressed.
func (s *SuppressNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if s.IsSuppressed() {
		return nil
	}
	return s.inner.Send(ctx, events)
}
