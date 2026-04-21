package alert

import "context"

// TransformFunc mutates or replaces events before forwarding.
type TransformFunc func(events []Event) []Event

// transformNotifier applies a TransformFunc to events before delegating.
type transformNotifier struct {
	transform TransformFunc
	next      Notifier
}

// NewTransformNotifier returns a Notifier that applies fn to each batch of
// events before passing them to next. If fn returns an empty slice the
// delegate is not called.
func NewTransformNotifier(fn TransformFunc, next Notifier) Notifier {
	return &transformNotifier{transform: fn, next: next}
}

func (t *transformNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	transformed := t.transform(events)
	if len(transformed) == 0 {
		return nil
	}
	return t.next.Send(ctx, transformed)
}

// PortTransform returns a TransformFunc that rewrites the Port field of every
// event using the provided mapping. Events whose port is not in the map are
// passed through unchanged.
func PortTransform(mapping map[uint16]uint16) TransformFunc {
	return func(events []Event) []Event {
		out := make([]Event, len(events))
		for i, e := range events {
			if mapped, ok := mapping[e.Entry.Port]; ok {
				e.Entry.Port = mapped
			}
			out[i] = e
		}
		return out
	}
}

// KindTransform returns a TransformFunc that overrides the Kind field of
// every event with the supplied value.
func KindTransform(kind string) TransformFunc {
	return func(events []Event) []Event {
		out := make([]Event, len(events))
		for i, e := range events {
			e.Kind = kind
			out[i] = e
		}
		return out
	}
}

// ChainTransforms composes multiple TransformFuncs left-to-right.
func ChainTransforms(fns ...TransformFunc) TransformFunc {
	return func(events []Event) []Event {
		for _, fn := range fns {
			events = fn(events)
			if len(events) == 0 {
				return nil
			}
		}
		return events
	}
}
