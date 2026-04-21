package alert

import (
	"testing"

	"github.com/user/portwatch/internal/scanner"
)

func TestHasKind_Match(t *testing.T) {
	events := []Event{makeRouteEvent("added", 80)}
	if !HasKind("added")(events) {
		t.Fatal("expected match")
	}
}

func TestHasKind_NoMatch(t *testing.T) {
	events := []Event{makeRouteEvent("removed", 80)}
	if HasKind("added")(events) {
		t.Fatal("expected no match")
	}
}

func TestHasPort_Match(t *testing.T) {
	events := []Event{makeRouteEvent("added", 443)}
	if !HasPort(443)(events) {
		t.Fatal("expected match")
	}
}

func TestHasPort_NoMatch(t *testing.T) {
	events := []Event{makeRouteEvent("added", 80)}
	if HasPort(443)(events) {
		t.Fatal("expected no match")
	}
}

func TestHasAddress_Match(t *testing.T) {
	events := []Event{{Kind: "added", Entry: scanner.Entry{Addr: "127.0.0.1", Port: 22, Proto: scanner.TCP}}}
	if !HasAddress("127.0.0.1")(events) {
		t.Fatal("expected match")
	}
}

func TestHasProtocol_Match(t *testing.T) {
	events := []Event{{Kind: "added", Entry: scanner.Entry{Addr: "0.0.0.0", Port: 53, Proto: scanner.UDP}}}
	if !HasProtocol(scanner.UDP)(events) {
		t.Fatal("expected match")
	}
}

func TestHasProtocol_NoMatch(t *testing.T) {
	events := []Event{{Kind: "added", Entry: scanner.Entry{Addr: "0.0.0.0", Port: 53, Proto: scanner.TCP}}}
	if HasProtocol(scanner.UDP)(events) {
		t.Fatal("expected no match")
	}
}

func TestAnyEvent_True(t *testing.T) {
	if !AnyEvent([]Event{makeRouteEvent("added", 1)}) {
		t.Fatal("expected true")
	}
}

func TestAnyEvent_Empty(t *testing.T) {
	if AnyEvent(nil) {
		t.Fatal("expected false for nil")
	}
}

func TestAllAdded_True(t *testing.T) {
	events := []Event{makeRouteEvent("added", 80), makeRouteEvent("added", 443)}
	if !AllAdded(events) {
		t.Fatal("expected true")
	}
}

func TestAllAdded_Mixed(t *testing.T) {
	events := []Event{makeRouteEvent("added", 80), makeRouteEvent("removed", 443)}
	if AllAdded(events) {
		t.Fatal("expected false for mixed events")
	}
}

func TestAllAdded_Empty(t *testing.T) {
	if AllAdded(nil) {
		t.Fatal("expected false for empty slice")
	}
}
