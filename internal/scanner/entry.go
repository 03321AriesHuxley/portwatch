package scanner

import "fmt"

// Entry represents a single bound port observed in /proc/net/tcp or tcp6.
type Entry struct {
	Proto     string
	LocalAddr string
	LocalPort uint16
	State     string
	Inode     uint64
}

// String returns a human-readable representation of the Entry.
func (e Entry) String() string {
	return fmt.Sprintf("%s:%s:%d", e.Proto, e.LocalAddr, e.LocalPort)
}

// Equal reports whether two entries represent the same port binding.
// Inode and State are intentionally excluded so that transient kernel
// state changes do not produce spurious diff events.
func (e Entry) Equal(other Entry) bool {
	return e.Proto == other.Proto &&
		e.LocalAddr == other.LocalAddr &&
		e.LocalPort == other.LocalPort
}

// Key returns a stable string key suitable for use in maps and dedup logic.
func (e Entry) Key() string {
	return e.String()
}

// IsListening reports whether the entry is in the LISTEN state.
// This is a convenience helper for filtering active listeners from a
// full snapshot that may include entries in other TCP states.
func (e Entry) IsListening() bool {
	return e.State == "LISTEN"
}
