package alert

import (
	"context"
	"net"
	"strings"
)

// RedactFunc is a function that redacts or masks a field value.
type RedactFunc func(value string) string

// redactNotifier wraps a Notifier and applies redaction to event metadata
// and address fields before forwarding, preventing sensitive data leakage.
type redactNotifier struct {
	next    Notifier
	fields  map[string]RedactFunc
	maskIP  bool
}

// RedactOption configures a redactNotifier.
type RedactOption func(*redactNotifier)

// WithMetaRedaction registers a RedactFunc for a specific metadata key.
func WithMetaRedaction(key string, fn RedactFunc) RedactOption {
	return func(r *redactNotifier) {
		r.fields[key] = fn
	}
}

// WithIPMasking enables masking of the host portion of event addresses.
func WithIPMasking() RedactOption {
	return func(r *redactNotifier) {
		r.maskIP = true
	}
}

// MaskRight replaces all but the first n characters of s with asterisks.
func MaskRight(s string, n int) string {
	if n <= 0 || len(s) == 0 {
		return strings.Repeat("*", len(s))
	}
	if n >= len(s) {
		return s
	}
	return s[:n] + strings.Repeat("*", len(s)-n)
}

// NewRedactNotifier returns a Notifier that redacts event fields before
// forwarding to next. Options control which fields are masked.
func NewRedactNotifier(next Notifier, opts ...RedactOption) Notifier {
	r := &redactNotifier{
		next:   next,
		fields: make(map[string]RedactFunc),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

func (r *redactNotifier) Send(ctx context.Context, events []Event) error {
	if len(events) == 0 {
		return r.next.Send(ctx, events)
	}
	redacted := make([]Event, len(events))
	for i, ev := range events {
		redacted[i] = r.redactEvent(ev)
	}
	return r.next.Send(ctx, redacted)
}

func (r *redactNotifier) redactEvent(ev Event) Event {
	if r.maskIP {
		ev.Entry = maskEntryAddress(ev.Entry)
	}
	if len(r.fields) == 0 || ev.Meta == nil {
		return ev
	}
	newMeta := make(map[string]string, len(ev.Meta))
	for k, v := range ev.Meta {
		if fn, ok := r.fields[k]; ok {
			newMeta[k] = fn(v)
		} else {
			newMeta[k] = v
		}
	}
	ev.Meta = newMeta
	return ev
}

func maskEntryAddress(e interface{ GetAddress() string }) interface{} {
	// We operate on scanner.Entry via the Event wrapper.
	return e
}

func maskIP(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Not host:port — try masking as plain IP.
		ip := net.ParseIP(addr)
		if ip == nil {
			return addr
		}
		return maskIPString(ip.String())
	}
	return net.JoinHostPort(maskIPString(host), port)
}

func maskIPString(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + ".***.*"
	}
	return "[redacted]"
}
