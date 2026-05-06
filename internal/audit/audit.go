// Package audit provides a simple append-only audit log that records
// every drift check result to a file for later inspection.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Entry is a single audit log record.
type Entry struct {
	Timestamp time.Time        `json:"timestamp"`
	Service   string           `json:"service"`
	Status    string           `json:"status"` // "match", "drifted", "missing"
	FilePath  string           `json:"file_path"`
	Ref       string           `json:"ref"`
	Details   string           `json:"details,omitempty"`
}

// Logger writes audit entries to an io.Writer as newline-delimited JSON.
type Logger struct {
	w io.Writer
}

// New returns a Logger that appends to the file at path, creating it if needed.
func New(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("audit: open %s: %w", path, err)
	}
	return &Logger{w: f}, nil
}

// NewWithWriter returns a Logger backed by the supplied writer (useful in tests).
func NewWithWriter(w io.Writer) *Logger {
	return &Logger{w: w}
}

// Record converts a drift.Report into audit entries and writes them.
func (l *Logger) Record(service, ref string, report *drift.Report) error {
	if report == nil {
		return nil
	}
	for _, ev := range report.Events {
		entry := Entry{
			Timestamp: time.Now().UTC(),
			Service:   service,
			Ref:       ref,
			FilePath:  ev.Path,
			Status:    string(ev.Status),
			Details:   ev.Detail,
		}
		if err := l.write(entry); err != nil {
			return err
		}
	}
	return nil
}

func (l *Logger) write(e Entry) error {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	return err
}
