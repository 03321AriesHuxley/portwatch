package alert_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

func makeMetricsEvent() alert.Event {
	return alert.Event{
		Kind:  alert.EventAdded,
		Entry: scanner.Entry{Addr: "0.0.0.0", Port: 9090, Protocol: "tcp"},
	}
}

type stubNotifier struct {
	err error
}

func (s *stubNotifier) Send(_ context.Context, events []alert.Event) error {
	if s.err != nil {
		return s.err
	}
	return nil
}

func TestMetricsNotifier_CountsSentEvents(t *testing.T) {
	inner := &stubNotifier{}
	mn := alert.NewMetricsNotifier(inner, nil)

	events := []alert.Event{makeMetricsEvent(), makeMetricsEvent()}
	if err := mn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := mn.Snapshot()
	if snap.TotalSent != 2 {
		t.Errorf("expected TotalSent=2, got %d", snap.TotalSent)
	}
}

func TestMetricsNotifier_CountsFailures(t *testing.T) {
	inner := &stubNotifier{err: errors.New("send failed")}
	mn := alert.NewMetricsNotifier(inner, nil)

	_ = mn.Send(context.Background(), []alert.Event{makeMetricsEvent()})

	snap := mn.Snapshot()
	if snap.TotalFailed != 1 {
		t.Errorf("expected TotalFailed=1, got %d", snap.TotalFailed)
	}
	if snap.TotalSent != 0 {
		t.Errorf("expected TotalSent=0, got %d", snap.TotalSent)
	}
}

func TestMetricsNotifier_CountsFiltered(t *testing.T) {
	inner := &stubNotifier{}
	mn := alert.NewMetricsNotifier(inner, nil)

	_ = mn.Send(context.Background(), []alert.Event{})
	_ = mn.Send(context.Background(), []alert.Event{})

	snap := mn.Snapshot()
	if snap.TotalFiltered != 2 {
		t.Errorf("expected TotalFiltered=2, got %d", snap.TotalFiltered)
	}
}

func TestMetricsNotifier_AcceptsExternalMetrics(t *testing.T) {
	m := &alert.Metrics{}
	inner := &stubNotifier{}
	mn := alert.NewMetricsNotifier(inner, m)

	_ = mn.Send(context.Background(), []alert.Event{makeMetricsEvent()})

	if m.TotalSent.Load() != 1 {
		t.Errorf("expected shared metrics TotalSent=1, got %d", m.TotalSent.Load())
	}
}
