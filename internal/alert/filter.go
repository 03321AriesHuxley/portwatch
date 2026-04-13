package alert

import "strings"

// Filter suppresses events whose port or address matches a configured
// exclusion list. It wraps an inner Notifier and forwards only the
// events that survive filtering.
type Filter struct {
	inner     Notifier
	excludePorts []uint16
	excludeAddrs []string
}

// NewFilter returns a Filter that wraps inner and drops any event whose
// port is in excludePorts or whose local address prefix matches any
// entry in excludeAddrs.
func NewFilter(inner Notifier, excludePorts []uint16, excludeAddrs []string) *Filter {
	return &Filter{
		inner:        inner,
		excludePorts: excludePorts,
		excludeAddrs: excludeAddrs,
	}
}

// Send filters events and forwards the remaining ones to the inner Notifier.
// Returns nil when all events are suppressed.
func (f *Filter) Send(events []Event) error {
	var kept []Event
	for _, e := range events {
		if f.excluded(e) {
			continue
		}
		kept = append(kept, e)
	}
	if len(kept) == 0 {
		return nil
	}
	return f.inner.Send(kept)
}

func (f *Filter) excluded(e Event) bool {
	for _, p := range f.excludePorts {
		if e.Entry.LocalPort == p {
			return true
		}
	}
	for _, addr := range f.excludeAddrs {
		if strings.HasPrefix(e.Entry.LocalAddr, addr) {
			return true
		}
	}
	return false
}
