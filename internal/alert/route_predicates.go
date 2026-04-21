package alert

// HasKind returns a predicate that matches events with the given kind (e.g. "added", "removed").
func HasKind(kind string) func(Event) bool {
	return func(e Event) bool {
		return e.Kind == kind
	}
}

// HasPort returns a predicate that matches events whose entry port equals port.
func HasPort(port uint16) func(Event) bool {
	return func(e Event) bool {
		return e.Entry.Port == port
	}
}

// HasAddress returns a predicate that matches events whose entry local address equals addr.
func HasAddress(addr string) func(Event) bool {
	return func(e Event) bool {
		return e.Entry.LocalAddr == addr
	}
}

// HasProtocol returns a predicate that matches events whose entry protocol equals proto.
func HasProtocol(proto string) func(Event) bool {
	return func(e Event) bool {
		return e.Entry.Protocol == proto
	}
}

// AnyEvent is a predicate that matches all events; useful as a catch-all route.
func AnyEvent(_ Event) bool { return true }

// Not negates a predicate.
func Not(pred func(Event) bool) func(Event) bool {
	return func(e Event) bool {
		return !pred(e)
	}
}

// All returns a predicate that is true only when all given predicates match.
func All(preds ...func(Event) bool) func(Event) bool {
	return func(e Event) bool {
		for _, p := range preds {
			if !p(e) {
				return false
			}
		}
		return true
	}
}

// Any returns a predicate that is true when at least one predicate matches.
func Any(preds ...func(Event) bool) func(Event) bool {
	return func(e Event) bool {
		for _, p := range preds {
			if p(e) {
				return true
			}
		}
		return false
	}
}

// HasMeta returns a predicate that matches events whose Meta map contains the given key.
func HasMeta(key string) func(Event) bool {
	return func(e Event) bool {
		_, ok := e.Meta[key]
		return ok
	}
}
