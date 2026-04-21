package alert

import (
	"context"
	"strings"
)

// Tagger is a function that returns a set of string tags for a given event.
type Tagger func(e Event) []string

// tagNotifier wraps a Notifier and attaches tags to each event's metadata
// before forwarding to the inner notifier.
type tagNotifier struct {
	inner  Notifier
	tagger Tagger
}

// TagOption configures a tagNotifier.
type TagOption func(*tagNotifier)

// WithTagger sets a custom tagger function.
func WithTagger(t Tagger) TagOption {
	return func(n *tagNotifier) {
		n.tagger = t
	}
}

// NewTagNotifier returns a Notifier that applies tags to each event before
// forwarding them to inner. Tags are appended to Event.Meta["tags"] as a
// comma-separated string.
func NewTagNotifier(inner Notifier, opts ...TagOption) Notifier {
	n := &tagNotifier{
		inner:  inner,
		tagger: defaultTagger,
	}
	for _, o := range opts {
		o(n)
	}
	return n
}

func (n *tagNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return n.inner.Send(ctx, events)
	}
	tagged := make([]Event, len(events))
	for i, e := range events {
		if e.Meta == nil {
			e.Meta = make(map[string]string)
		}
		tags := n.tagger(e)
		if len(tags) > 0 {
			e.Meta["tags"] = strings.Join(tags, ",")
		}
		tagged[i] = e
	}
	return n.inner.Send(ctx, tagged)
}

// defaultTagger derives tags from the event kind and protocol.
func defaultTagger(e Event) []string {
	tags := []string{string(e.Kind)}
	if e.Entry.Protocol != "" {
		tags = append(tags, e.Entry.Protocol)
	}
	return tags
}
