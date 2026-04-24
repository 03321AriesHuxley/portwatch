package alert

import "context"

// Labeler is a function that returns a map of static or dynamic labels
// to attach to every event in a batch.
type Labeler func(e Event) map[string]string

// labelNotifier wraps a Notifier and attaches labels to each event's Meta map.
type labelNotifier struct {
	next    Notifier
	labeler Labeler
}

// WithLabeler returns a Labeler that always returns the given static labels.
func WithLabeler(labels map[string]string) Labeler {
	return func(_ Event) map[string]string {
		out := make(map[string]string, len(labels))
		for k, v := range labels {
			out[k] = v
		}
		return out
	}
}

// NewLabelNotifier returns a Notifier that merges labels produced by labeler
// into each event's Meta field before forwarding to next.
// Existing meta keys are NOT overwritten; labeler-provided keys only fill gaps.
func NewLabelNotifier(next Notifier, labeler Labeler) Notifier {
	return &labelNotifier{next: next, labeler: labeler}
}

func (l *labelNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}

	labeled := make([]Event, len(events))
	for i, e := range events {
		extra := l.labeler(e)
		if len(extra) == 0 {
			labeled[i] = e
			continue
		}
		merged := make(map[string]string, len(e.Meta)+len(extra))
		for k, v := range extra {
			merged[k] = v
		}
		// existing meta wins over labeler values
		for k, v := range e.Meta {
			merged[k] = v
		}
		e.Meta = merged
		labeled[i] = e
	}

	return l.next.Send(ctx, labeled)
}
