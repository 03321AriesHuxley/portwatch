package alert

import (
	"context"
)

// RoutePredicate is a function that decides whether a set of events
// should be forwarded to a particular notifier branch.
type RoutePredicate func(events []Event) bool

// route is a single branch in a Router.
type route struct {
	predicate RoutePredicate
	notifier  Notifier
}

// Router dispatches events to one or more notifiers based on predicates.
// Events are evaluated against each route in order; all matching routes
// receive the events (fan-out). If no route matches, events are dropped
// unless a fallback notifier has been configured.
type Router struct {
	routes   []route
	fallback Notifier
}

// RouterOption configures a Router.
type RouterOption func(*Router)

// WithFallback sets a notifier that receives events when no route matches.
func WithFallback(n Notifier) RouterOption {
	return func(r *Router) {
		r.fallback = n
	}
}

// NewRouter creates a Router with the provided routes and options.
// Each route is a (predicate, notifier) pair.
func NewRouter(routes []route, opts ...RouterOption) *Router {
	r := &Router{routes: routes}
	for _, o := range opts {
		o(r)
	}
	return r
}

// AddRoute appends a route to the router at runtime.
func (r *Router) AddRoute(pred RoutePredicate, n Notifier) {
	r.routes = append(r.routes, route{predicate: pred, notifier: n})
}

// Send evaluates each route predicate and forwards events to all matching
// notifiers. The first error encountered is returned; remaining notifiers
// are still attempted.
func (r *Router) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	matched := false
	var firstErr error
	for _, rt := range r.routes {
		if rt.predicate(events) {
			matched = true
			if err := rt.notifier.Send(ctx, events); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	if !matched && r.fallback != nil {
		if err := r.fallback.Send(ctx, events); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
