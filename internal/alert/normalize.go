package alert

import (
	"context"
	"net"
	"strings"
)

// NormalizeFunc transforms an event's fields into a canonical form.
type NormalizeFunc func(Event) Event

// normalizeNotifier applies a normalization function to all events before
// forwarding them to the inner Notifier.
type normalizeNotifier struct {
	inner     Notifier
	normalize NormalizeFunc
}

// NewNormalizeNotifier returns a Notifier that applies fn to each event before
// delegating to inner. If fn is nil, a default normalizer is used that
// lower-cases the address and expands unspecified IPs to "0.0.0.0" / "::".
func NewNormalizeNotifier(inner Notifier, fn NormalizeFunc) Notifier {
	if fn == nil {
		fn = DefaultNormalize
	}
	return &normalizeNotifier{inner: inner, normalize: fn}
}

func (n *normalizeNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return nil
	}
	normalized := make([]Event, len(events))
	for i, e := range events {
		normalized[i] = n.normalize(e)
	}
	return n.inner.Send(ctx, normalized)
}

// DefaultNormalize is the built-in normalization function. It:
//   - lower-cases the address string
//   - replaces blank / unspecified addresses with "0.0.0.0" (IPv4) or "::" (IPv6)
//   - trims whitespace from the protocol field
func DefaultNormalize(e Event) Event {
	addr := strings.TrimSpace(strings.ToLower(e.Entry.Addr))
	if addr == "" {
		addr = "0.0.0.0"
	} else {
		ip := net.ParseIP(addr)
		if ip != nil && ip.IsUnspecified() {
			if ip.To4() != nil {
				addr = "0.0.0.0"
			} else {
				addr = "::"
			}
		}
	}
	e.Entry.Addr = addr
	e.Entry.Proto = strings.TrimSpace(strings.ToLower(e.Entry.Proto))
	return e
}
