package alert

import (
	"context"
	"math/rand"
	"sync"
)

// SamplingNotifier forwards events to the inner Notifier with a given
// probability in the range (0.0, 1.0]. A rate of 1.0 forwards everything.
type SamplingNotifier struct {
	mu    sync.Mutex
	inner Notifier
	rate  float64
	rng   *rand.Rand
}

// NewSamplingNotifier creates a SamplingNotifier that samples events at the
// given rate. rate must be in (0, 1]; values outside that range are clamped.
func NewSamplingNotifier(inner Notifier, rate float64, src rand.Source) *SamplingNotifier {
	if rate <= 0 {
		rate = 0.01
	}
	if rate > 1 {
		rate = 1
	}
	if src == nil {
		src = rand.NewSource(42)
	}
	return &SamplingNotifier{
		inner: inner,
		rate:  rate,
		rng:   rand.New(src),
	}
}

// Send forwards a sampled subset of events to the inner notifier.
// If no events pass the sampler, Send returns nil without calling inner.
func (s *SamplingNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	s.mu.Lock()
	sampled := make([]Event, 0, len(events))
	for _, e := range events {
		if s.rng.Float64() < s.rate {
			sampled = append(sampled, e)
		}
	}
	s.mu.Unlock()

	if len(sampled) == 0 {
		return nil
	}
	return s.inner.Send(ctx, sampled)
}
