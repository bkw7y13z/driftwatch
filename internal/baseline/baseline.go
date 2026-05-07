// Package baseline manages pinned baseline snapshots for drift comparison.
// A baseline represents a known-good state captured at a specific git ref,
// allowing operators to compare live config against a pinned reference rather
// than always HEAD.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Entry holds a pinned baseline for a single service.
type Entry struct {
	Service   string            `json:"service"`
	Ref       string            `json:"ref"`
	PinnedAt  time.Time         `json:"pinned_at"`
	Files     map[string]string `json:"files"` // path -> content hash
}

// Store persists and retrieves baseline entries.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	path    string
}

// New returns a Store backed by the given file path.
// If the file exists it is loaded immediately.
func New(path string) (*Store, error) {
	s := &Store{
		entries: make(map[string]*Entry),
		path:    path,
	}
	if _, err := os.Stat(path); err == nil {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("baseline: load %s: %w", path, err)
		}
	}
	return s, nil
}

// Pin records a baseline entry for the given service, replacing any prior entry.
func (s *Store) Pin(e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.PinnedAt = time.Now().UTC()
	s.entries[e.Service] = &e
	return s.save()
}

// Get returns the baseline entry for a service, or false if none exists.
func (s *Store) Get(service string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[service]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// Remove deletes the baseline for the given service and persists the change.
func (s *Store) Remove(service string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, service)
	return s.save()
}

func (s *Store) save() error {
	f, err := os.CreateTemp("", "baseline-*.json")
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(s.entries); err != nil {
		f.Close()
		return err
	}
	f.Close()
	return os.Rename(f.Name(), s.path)
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&s.entries)
}
