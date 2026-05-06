// Package watcher monitors file paths on disk and compares them against
// their declared state stored in a git repository.
package watcher

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/yourorg/driftwatch/internal/drift"
)

// FileState holds the raw bytes of a file read from disk.
type FileState struct {
	Path    string
	Content []byte
	Missing bool
}

// Watcher reads live files from the local filesystem.
type Watcher struct {
	basePaths []string
	log       *slog.Logger
}

// New creates a Watcher that resolves files relative to the given base paths.
func New(basePaths []string, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.Default()
	}
	return &Watcher{basePaths: basePaths, log: log}
}

// ReadFile reads the content of the file at the given relative path,
// searching each base path in order. Returns FileState with Missing=true
// if the file cannot be found in any base path.
func (w *Watcher) ReadFile(_ context.Context, relPath string) FileState {
	for _, base := range w.basePaths {
		full := filepath.Join(base, relPath)
		data, err := os.ReadFile(full)
		if err == nil {
			w.log.Debug("watcher: read file", "path", full)
			return FileState{Path: full, Content: data}
		}
	}
	w.log.Debug("watcher: file not found", "relPath", relPath)
	return FileState{Path: relPath, Missing: true}
}

// ReadAll reads every path in relPaths and returns a map keyed by the
// relative path. Errors are captured as Missing entries.
func (w *Watcher) ReadAll(ctx context.Context, relPaths []string) (map[string]FileState, error) {
	result := make(map[string]FileState, len(relPaths))
	for _, p := range relPaths {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("watcher: context cancelled: %w", ctx.Err())
		default:
		}
		result[p] = w.ReadFile(ctx, p)
	}
	return result, nil
}

// ToLiveContent converts a FileState map to the map[string][]byte format
// expected by drift.Detector.
func ToLiveContent(states map[string]FileState) map[string][]byte {
	out := make(map[string][]byte, len(states))
	for k, s := range states {
		if !s.Missing {
			out[k] = s.Content
		}
	}
	return out
}

// MissingPaths returns the relative paths whose FileState has Missing=true.
func MissingPaths(states map[string]FileState) []string {
	var missing []string
	for k, s := range states {
		if s.Missing {
			missing = append(missing, k)
		}
	}
	_ = drift.StatusMatch // ensure drift package linkage is valid
	return missing
}
