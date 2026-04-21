package alert

import (
	"context"
	"sync"
	"time"
)

// EscalationPolicy defines when and how to escalate unacknowledged events.
type EscalationPolicy struct {
	// After defines how long to wait before escalating.
	After time.Duration
	// Target is the notifier to escalate to.
	Target Notifier
}

// EscalateNotifier forwards events to a primary notifier and escalates
// to a secondary notifier if the primary fails or events go unacknowledged
// within the policy window.
type EscalateNotifier struct {
	primary  Notifier
	policies []EscalationPolicy
	clock    func() time.Time
	mu       sync.Mutex
	lastFail time.Time
}

// NewEscalateNotifier creates a notifier that escalates to fallback targets
// according to the provided policies when the primary notifier fails.
func NewEscalateNotifier(primary Notifier, policies ...EscalationPolicy) *EscalateNotifier {
	return &EscalateNotifier{
		primary:  primary,
		policies: policies,
		clock:    time.Now,
	}
}

// Send attempts the primary notifier. On failure, it evaluates each escalation
// policy in order and forwards to the first whose delay window has elapsed.
func (e *EscalateNotifier) Send(ctx context.Context, events []Event) error {
	err := e.primary.Send(ctx, events)
	if err == nil {
		e.mu.Lock()
		e.lastFail = time.Time{}
		e.mu.Unlock()
		return nil
	}

	now := e.clock()
	e.mu.Lock()
	if e.lastFail.IsZero() {
		e.lastFail = now
	}
	failSince := e.lastFail
	e.mu.Unlock()

	for _, p := range e.policies {
		if now.Sub(failSince) >= p.After {
			if escalateErr := p.Target.Send(ctx, events); escalateErr == nil {
				return nil
			}
		}
	}

	return err
}
