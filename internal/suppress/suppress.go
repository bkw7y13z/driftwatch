// Package suppress provides a mechanism to temporarily silence drift
// alerts for known-good exceptions or scheduled maintenance windows.
package suppress

import (
	"sync"
	"time"
)

// Entry represents a single suppression rule.
type Entry struct {
	Service  string
	Path     string
	Expiry   time.Time
	Reason   string
}

// Store holds active suppression entries.
type Store struct {
	mu      sync.RWMutex
	entries []Entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{now: time.Now}
}

// Add registers a suppression entry that expires after ttl.
func (s *Store) Add(service, path, reason string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{
		Service: service,
		Path:    path,
		Expiry:  s.now().Add(ttl),
		Reason:  reason,
	})
}

// IsSuppressed reports whether alerts for the given service+path pair
// should be silenced at the current moment.
func (s *Store) IsSuppressed(service, path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	for _, e := range s.entries {
		if e.Service == service && e.Path == path && now.Before(e.Expiry) {
			return true
		}
	}
	return false
}

// Purge removes all expired entries. It is safe to call periodically.
func (s *Store) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	active := s.entries[:0]
	for _, e := range s.entries {
		if now.Before(e.Expiry) {
			active = append(active, e)
		}
	}
	s.entries = active
}

// Active returns a snapshot of all currently active (non-expired) entries.
func (s *Store) Active() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.now()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if now.Before(e.Expiry) {
			out = append(out, e)
		}
	}
	return out
}
