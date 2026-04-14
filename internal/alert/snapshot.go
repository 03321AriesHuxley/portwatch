package alert

import (
	"sync"
	"time"
)

// NotifierSnapshot holds a point-in-time view of notifier health metrics.
type NotifierSnapshot struct {
	Name      string
	Sent      int64
	Failed    int64
	Filtered  int64
	LastSent  time.Time
	LastError error
}

// SnapshotStore collects and retrieves named notifier snapshots.
type SnapshotStore struct {
	mu        sync.RWMutex
	snapshots map[string]*NotifierSnapshot
}

// NewSnapshotStore creates an empty SnapshotStore.
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string]*NotifierSnapshot),
	}
}

// Record upserts a snapshot for the given notifier name.
func (s *SnapshotStore) Record(snap NotifierSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.snapshots[snap.Name] = &snap
}

// Get retrieves a snapshot by name. Returns false if not found.
func (s *SnapshotStore) Get(name string) (NotifierSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.snapshots[name]
	if !ok {
		return NotifierSnapshot{}, false
	}
	return *snap, true
}

// All returns a copy of all stored snapshots.
func (s *SnapshotStore) All() []NotifierSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NotifierSnapshot, 0, len(s.snapshots))
	for _, snap := range s.snapshots {
		out = append(out, *snap)
	}
	return out
}
