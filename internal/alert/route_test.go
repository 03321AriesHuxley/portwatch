package alert

import (
	"context"
	"errors"
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func makeRouteEvent(kind string, port uint16) Event {
	return Event{
		Kind:  kind,
		Entry: scanner.Entry{Port: port, Addr: "0.0.0.0", Proto: scanner.TCP},
	}
}

func TestRouter_NoRoutes_DropsEvents(t *testing.T) {
	r := NewRouter(nil)
	err := r.Send(context.Background(), []Event{makeRouteEvent("added", 80)})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRouter_FallbackReceivesUnmatched(t *testing.T) {
	var got []Event
	fb := NotifierFunc(func(_ context.Context, evs []Event) error {
		got = evs
		return nil
	})
	r := NewRouter(nil, WithFallback(fb))
	events := []Event{makeRouteEvent("added", 443)}
	if err := r.Send(context.Background(), events); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
}

func TestRouter_MatchingRouteCalled(t *testing.T) {
	var called bool
	n := NotifierFunc(func(_ context.Context, _ []Event) error {
		called = true
		return nil
	})
	routes := []route{{predicate: HasKind("added"), notifier: n}}
	r := NewRouter(routes)
	if err := r.Send(context.Background(), []Event{makeRouteEvent("added", 22)}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected notifier to be called")
	}
}

func TestRouter_NonMatchingRouteSkipped(t *testing.T) {
	var called bool
	n := NotifierFunc(func(_ context.Context, _ []Event) error {
		called = true
		return nil
	})
	routes := []route{{predicate: HasKind("removed"), notifier: n}}
	r := NewRouter(routes)
	if err := r.Send(context.Background(), []Event{makeRouteEvent("added", 22)}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("expected notifier NOT to be called")
	}
}

func TestRouter_MultipleMatchingRoutes_AllCalled(t *testing.T) {
	counts := make([]int, 2)
	makeN := func(i int) Notifier {
		return NotifierFunc(func(_ context.Context, _ []Event) error {
			counts[i]++
			return nil
		})
	}
	routes := []route{
		{predicate: AnyEvent, notifier: makeN(0)},
		{predicate: HasKind("added"), notifier: makeN(1)},
	}
	r := NewRouter(routes)
	if err := r.Send(context.Background(), []Event{makeRouteEvent("added", 8080)}); err != nil {
		t.Fatal(err)
	}
	for i, c := range counts {
		if c != 1 {
			t.Errorf("route %d: expected 1 call, got %d", i, c)
		}
	}
}

func TestRouter_ErrorPropagated(t *testing.T) {
	sentinel := errors.New("boom")
	n := NotifierFunc(func(_ context.Context, _ []Event) error { return sentinel })
	routes := []route{{predicate: AnyEvent, notifier: n}}
	r := NewRouter(routes)
	err := r.Send(context.Background(), []Event{makeRouteEvent("added", 9000)})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestRouter_EmptyEvents_Noop(t *testing.T) {
	var called bool
	n := NotifierFunc(func(_ context.Context, _ []Event) error {
		called = true
		return nil
	})
	routes := []route{{predicate: AnyEvent, notifier: n}}
	r := NewRouter(routes)
	if err := r.Send(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("expected notifier NOT to be called for empty events")
	}
}
