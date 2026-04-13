package alert_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/scanner"
)

type fakeNotifier struct {
	called int
	err    error
}

func (f *fakeNotifier) Send(_ context.Context, events []alert.PortEvent) error {
	f.called++
	return f.err
}

func sampleEvents() []alert.PortEvent {
	return []alert.PortEvent{
		{
			Type:       alert.EventOpened,
			Entry:      scanner.Entry{LocalAddress: "0.0.0.0:8080"},
			DetectedAt: time.Now(),
		},
	}
}

func TestDispatcher_NoNotifiers(t *testing.T) {
	d := alert.NewDispatcher()
	errs := d.Dispatch(context.Background(), sampleEvents())
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestDispatcher_EmptyEvents(t *testing.T) {
	n := &fakeNotifier{}
	d := alert.NewDispatcher(n)
	errs := d.Dispatch(context.Background(), nil)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for empty events")
	}
	if n.called != 0 {
		t.Errorf("notifier should not be called for empty events")
	}
}

func TestDispatcher_CallsAllNotifiers(t *testing.T) {
	a, b := &fakeNotifier{}, &fakeNotifier{}
	d := alert.NewDispatcher(a, b)
	errs := d.Dispatch(context.Background(), sampleEvents())
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if a.called != 1 || b.called != 1 {
		t.Errorf("expected each notifier called once, got a=%d b=%d", a.called, b.called)
	}
}

func TestDispatcher_CollectsErrors(t *testing.T) {
	sentinel := errors.New("send failed")
	good := &fakeNotifier{}
	bad := &fakeNotifier{err: sentinel}
	d := alert.NewDispatcher(good, bad)
	errs := d.Dispatch(context.Background(), sampleEvents())
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if !errors.Is(errs[0], sentinel) {
		t.Errorf("unexpected error: %v", errs[0])
	}
	if good.called != 1 {
		t.Errorf("good notifier should still have been called")
	}
}
