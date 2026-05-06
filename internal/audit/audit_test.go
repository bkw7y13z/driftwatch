package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/audit"
	"github.com/yourorg/driftwatch/internal/drift"
)

func makeReport(events []drift.Event) *drift.Report {
	return &drift.Report{Events: events}
}

func TestRecord_NilReport(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)
	if err := l.Record("svc", "main", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for nil report, got %q", buf.String())
	}
}

func TestRecord_EmptyReport(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)
	if err := l.Record("svc", "main", makeReport(nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty report")
	}
}

func TestRecord_WritesOneLinePerEvent(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	reprot := makeReport([]drift.Event{
		{Path: "/etc/app.conf", Status: drift.StatusMatch},
		{Path: "/etc/db.conf", Status: drift.StatusDrifted, Detail: "hash mismatch"},
	})

	if err := l.Record("my-service", "v1.2.3", reprot); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestRecord_EntryFields(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewWithWriter(&buf)

	reprot := makeReport([]drift.Event{
		{Path: "/etc/app.conf", Status: drift.StatusDrifted, Detail: "hash mismatch"},
	})

	before := time.Now().UTC()
	if err := l.Record("my-service", "abc123", reprot); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := time.Now().UTC()

	var entry audit.Entry
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entry); err != nil {
		t.Fatalf("failed to unmarshal entry: %v", err)
	}

	if entry.Service != "my-service" {
		t.Errorf("service: want %q got %q", "my-service", entry.Service)
	}
	if entry.Ref != "abc123" {
		t.Errorf("ref: want %q got %q", "abc123", entry.Ref)
	}
	if entry.FilePath != "/etc/app.conf" {
		t.Errorf("file_path: want %q got %q", "/etc/app.conf", entry.FilePath)
	}
	if entry.Status != "drifted" {
		t.Errorf("status: want %q got %q", "drifted", entry.Status)
	}
	if entry.Details != "hash mismatch" {
		t.Errorf("details: want %q got %q", "hash mismatch", entry.Details)
	}
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("timestamp %v out of expected range [%v, %v]", entry.Timestamp, before, after)
	}
}

func TestNew_BadPath(t *testing.T) {
	_, err := audit.New("/nonexistent-dir/audit.log")
	if err == nil {
		t.Fatal("expected error for bad path, got nil")
	}
}
