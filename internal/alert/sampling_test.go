package alert_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/deanrtaylor1/portwatch/internal/alert"
	"github.com/deanrtaylor1/portwatch/internal/scanner"
)

func makeSamplingEvents(n int) []alert.Event {
	events := make([]alert.Event, n)
	for i := range events {
		events[i] = alert.Event{
			Kind:  alert.EventAdded,
			Entry: scanner.Entry{Port: uint16(8000 + i), Protocol: "tcp"},
		}
	}
	return events
}

func TestSamplingNotifier_RateOne_ForwardsAll(t *testing.T) {
	captured := &captureNotifier{}
	src := rand.NewSource(1)
	sn := alert.NewSamplingNotifier(captured, 1.0, src)

	events := makeSamplingEvents(10)
	if err := sn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.events) != 10 {
		t.Fatalf("expected 10 events forwarded, got %d", len(captured.events))
	}
}

func TestSamplingNotifier_RateZeroClampedToMin(t *testing.T) {
	captured := &captureNotifier{}
	src := rand.NewSource(99)
	// rate 0 should be clamped to 0.01 — with only 5 events it is very likely
	// none pass, but we just verify no panic and inner may or may not be called.
	sn := alert.NewSamplingNotifier(captured, 0.0, src)
	if err := sn.Send(context.Background(), makeSamplingEvents(5)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSamplingNotifier_EmptyEvents(t *testing.T) {
	captured := &captureNotifier{}
	sn := alert.NewSamplingNotifier(captured, 1.0, nil)
	if err := sn.Send(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured.events) != 0 {
		t.Fatalf("expected no events forwarded for empty input")
	}
}

func TestSamplingNotifier_RateHalf_ReducesEvents(t *testing.T) {
	captured := &captureNotifier{}
	src := rand.NewSource(7)
	sn := alert.NewSamplingNotifier(captured, 0.5, src)

	events := makeSamplingEvents(1000)
	if err := sn.Send(context.Background(), events); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With seed 7 and rate 0.5 over 1000 events we expect roughly 500 ± 100.
	got := len(captured.events)
	if got < 350 || got > 650 {
		t.Fatalf("expected ~500 sampled events, got %d", got)
	}
}
