package alert

import (
	"context"
	"sync"
	"time"
)

// MuteNotifier wraps a Notifier and suppresses all sends during an active mute
// window. Muting can be activated for a fixed duration or cancelled early.
type MuteNotifier struct {
	inner    Notifier
	mu       sync.RWMutex
	muteUntil time.Time
}

// NewMuteNotifier returns a MuteNotifier wrapping inner.
func NewMuteNotifier(inner Notifier) *MuteNotifier {
	return &MuteNotifier{inner: inner}
}

// Mute silences the notifier for the given duration.
func (m *MuteNotifier) Mute(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muteUntil = time.Now().Add(d)
}

// Unmute cancels any active mute immediately.
func (m *MuteNotifier) Unmute() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muteUntil = time.Time{}
}

// IsMuted reports whether the notifier is currently muted.
func (m *MuteNotifier) IsMuted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Now().Before(m.muteUntil)
}

// Send forwards events to the inner notifier unless a mute window is active.
// If muted, Send returns nil without calling the inner notifier.
func (m *MuteNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	if m.IsMuted() {
		return nil
	}
	return m.inner.Send(ctx, events)
}
