package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry holds a single file's hash and metadata at a point in time.
type Entry struct {
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	Ref       string    `json:"ref"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Snapshot is a collection of file entries for a service.
type Snapshot struct {
	ServiceName string           `json:"service_name"`
	Entries     map[string]Entry `json:"entries"`
	CreatedAt   time.Time        `json:"created_at"`
}

// New creates an empty Snapshot for the given service.
func New(serviceName string) *Snapshot {
	return &Snapshot{
		ServiceName: serviceName,
		Entries:     make(map[string]Entry),
		CreatedAt:   time.Now().UTC(),
	}
}

// Set records or updates an entry in the snapshot.
func (s *Snapshot) Set(path, hash, ref string) {
	s.Entries[path] = Entry{
		Path:       path,
		Hash:       hash,
		Ref:        ref,
		RecordedAt: time.Now().UTC(),
	}
}

// Get retrieves an entry by path. Returns false if not found.
func (s *Snapshot) Get(path string) (Entry, bool) {
	e, ok := s.Entries[path]
	return e, ok
}

// Save persists the snapshot as JSON to dir/<serviceName>.snapshot.json.
func (s *Snapshot) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("snapshot: create dir: %w", err)
	}
	fileName := filepath.Join(dir, s.ServiceName+".snapshot.json")
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// Load reads a previously saved snapshot from dir/<serviceName>.snapshot.json.
func Load(dir, serviceName string) (*Snapshot, error) {
	fileName := filepath.Join(dir, serviceName+".snapshot.json")
	f, err := os.Open(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no prior snapshot is not an error
		}
		return nil, fmt.Errorf("snapshot: open: %w", err)
	}
	defer f.Close()
	var snap Snapshot
	if err := json.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}
	return &snap, nil
}
