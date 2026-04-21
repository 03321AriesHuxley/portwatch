package alert

import "github.com/user/portwatch/internal/scanner"

// HasKind returns a RoutePredicate that matches when at least one event
// has the specified kind ("added" or "removed").
func HasKind(kind string) RoutePredicate {
	return func(events []Event) bool {
		for _, e := range events {
			if e.Kind == kind {
				return true
			}
		}
		return false
	}
}

// HasPort returns a RoutePredicate that matches when at least one event
// involves the given port number.
func HasPort(port uint16) RoutePredicate {
	return func(events []Event) bool {
		for _, e := range events {
			if e.Entry.Port == port {
				return true
			}
		}
		return false
	}
}

// HasAddress returns a RoutePredicate that matches when at least one event
// involves the given local address string.
func HasAddress(addr string) RoutePredicate {
	return func(events []Event) bool {
		for _, e := range events {
			if e.Entry.Addr == addr {
				return true
			}
		}
		return false
	}
}

// HasProtocol returns a RoutePredicate that matches when at least one event
// involves the given protocol.
func HasProtocol(proto scanner.Protocol) RoutePredicate {
	return func(events []Event) bool {
		for _, e := range events {
			if e.Entry.Proto == proto {
				return true
			}
		}
		return false
	}
}

// AnyEvent is a RoutePredicate that always matches (catch-all).
func AnyEvent(events []Event) bool {
	return len(events) > 0
}

// AllAdded returns true only when every event in the slice is an "added" event.
func AllAdded(events []Event) bool {
	if len(events) == 0 {
		return false
	}
	for _, e := range events {
		if e.Kind != "added" {
			return false
		}
	}
	return true
}
