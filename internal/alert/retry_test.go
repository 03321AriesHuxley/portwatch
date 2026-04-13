package alert_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

type countingNotifier struct {
	calls  atomic.Int32
	failN  int // fail the first N calls
	err    error
}

func (c *countingNotifier) Send(_ context.Context, _ []alert.Event) error {
	n := int(c.calls.Add(1))
	if n <= c.failN {
		return c.err
	}
	return nil
}

func TestRetryNotifier_SucceedsFirstAttempt(t *testing.T) {
	inner := &countingNotifier{failN: 0}
	r := alert.NewRetryNotifier(inner, 3, time.Millisecond)
	err := r.Send(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if inner.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls.Load())
	}
}

func TestRetryNotifier_RetriesAndSucceeds(t *testing.T) {
	sentinel := errors.New("temporary")
	inner := &countingNotifier{failN: 2, err: sentinel}
	r := alert.NewRetryNotifier(inner, 3, time.Millisecond)
	err := r.Send(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected nil error after retry, got %v", err)
	}
	if inner.calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls.Load())
	}
}

func TestRetryNotifier_ExhaustsAttempts(t *testing.T) {
	sentinel := errors.New("permanent")
	inner := &countingNotifier{failN: 5, err: sentinel}
	r := alert.NewRetryNotifier(inner, 3, time.Millisecond)
	err := r.Send(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel wrapped in error, got %v", err)
	}
	if inner.calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls.Load())
	}
}

func TestRetryNotifier_RespectsContextCancel(t *testing.T) {
	sentinel := errors.New("fail")
	inner := &countingNotifier{failN: 10, err: sentinel}
	r := alert.NewRetryNotifier(inner, 5, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := r.Send(ctx, nil)
	if err == nil {
		t.Fatal("expected context error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
