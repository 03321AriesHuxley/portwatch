package alert

import (
	"context"
	"sync/atomic"
)

// Metrics tracks send statistics for a Notifier.
type Metrics struct {
	TotalSent     atomic.Int64
	TotalFailed   atomic.Int64
	TotalFiltered atomic.Int64
}

// MetricsNotifier wraps a Notifier and records send metrics.
type MetricsNotifier struct {
	inner   Notifier
	metrics *Metrics
}

// NewMetricsNotifier returns a MetricsNotifier wrapping inner.
func NewMetricsNotifier(inner Notifier, m *Metrics) *MetricsNotifier {
	if m == nil {
		m = &Metrics{}
	}
	return &MetricsNotifier{inner: inner, metrics: m}
}

// Send forwards events to the inner Notifier and records outcomes.
func (mn *MetricsNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		mn.metrics.TotalFiltered.Add(1)
		return nil
	}
	err := mn.inner.Send(ctx, events)
	if err != nil {
		mn.metrics.TotalFailed.Add(1)
		return err
	}
	mn.metrics.TotalSent.Add(int64(len(events)))
	return nil
}

// Snapshot returns a point-in-time copy of the current metric values.
func (mn *MetricsNotifier) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalSent:     mn.metrics.TotalSent.Load(),
		TotalFailed:   mn.metrics.TotalFailed.Load(),
		TotalFiltered: mn.metrics.TotalFiltered.Load(),
	}
}

// MetricsSnapshot is an immutable copy of Metrics values.
type MetricsSnapshot struct {
	TotalSent     int64
	TotalFailed   int64
	TotalFiltered int64
}
