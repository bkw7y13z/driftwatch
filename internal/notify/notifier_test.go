package notify

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeReport(results []drift.Result) *drift.Report {
	return &drift.Report{Results: results}
}

func TestNotify_NoEvents_OnAllMatch(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)

	report := makeReport([]drift.Result{
		{Path: "config.yaml", Status: drift.StatusMatch},
	})

	count, err := n.Notify("svc-a", report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 events, got %d", count)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestNotify_WarnOnDrifted(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)

	report := makeReport([]drift.Result{
		{Path: "app.conf", Status: drift.StatusDrifted},
	})

	count, err := n.Notify("svc-b", report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
	out := buf.String()
	if !strings.Contains(out, "[WARN]") {
		t.Errorf("expected WARN level in output, got: %s", out)
	}
	if !strings.Contains(out, "app.conf") {
		t.Errorf("expected file path in output, got: %s", out)
	}
	if !strings.Contains(out, "service=svc-b") {
		t.Errorf("expected service name in output, got: %s", out)
	}
}

func TestNotify_ErrorOnMissing(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)

	report := makeReport([]drift.Result{
		{Path: "secrets.env", Status: drift.StatusMissing},
	})

	count, err := n.Notify("svc-c", report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
	out := buf.String()
	if !strings.Contains(out, "[ERROR]") {
		t.Errorf("expected ERROR level in output, got: %s", out)
	}
}

func TestNotify_NilReport(t *testing.T) {
	n := New(nil)
	_, err := n.Notify("svc-d", nil)
	if err == nil {
		t.Error("expected error for nil report, got nil")
	}
}

func TestNew_DefaultsToStdout(t *testing.T) {
	n := New(nil)
	if n.out == nil {
		t.Error("expected non-nil writer when nil passed to New")
	}
}
