package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                       // blocking calls
	CircuitHalfOpen                   // testing recovery
)

// CircuitBreaker wraps a Notifier and stops forwarding events when the
// downstream notifier fails repeatedly, reopening after a cooldown period.
type CircuitBreaker struct {
	mu           sync.Mutex
	inner        Notifier
	maxFailures  int
	cooldown     time.Duration
	failures     int
	state        CircuitState
	openedAt     time.Time
	now          func() time.Time
}

// NewCircuitBreaker returns a Notifier that trips open after maxFailures
// consecutive errors from inner, and resets after cooldown.
func NewCircuitBreaker(inner Notifier, maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		inner:       inner,
		maxFailures: maxFailures,
		cooldown:    cooldown,
		state:       CircuitClosed,
		now:         time.Now,
	}
}

// Send forwards events to the inner Notifier unless the circuit is open.
func (cb *CircuitBreaker) Send(ctx context.Context, events []Event) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitOpen:
		if cb.now().Sub(cb.openedAt) < cb.cooldown {
			return fmt.Errorf("circuit open: too many consecutive failures")
		}
		// transition to half-open to allow a probe
		cb.state = CircuitHalfOpen
	case CircuitClosed, CircuitHalfOpen:
		// proceed
	}

	err := cb.inner.Send(ctx, events)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = CircuitOpen
			cb.openedAt = cb.now()
		}
		return err
	}

	// success — reset
	cb.failures = 0
	cb.state = CircuitClosed
	return nil
}

// State returns the current CircuitState (safe for concurrent use).
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
