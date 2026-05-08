package diffrender_test

import (
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/diffrender"
	"github.com/example/driftwatch/internal/drift"
)

func makeEvent(service, path string, status drift.Status, expected, actual string) drift.Event {
	return drift.Event{
		Service:  service,
		Path:     path,
		Status:   status,
		Expected: expected,
		Actual:   actual,
	}
}

func TestRender_NilReport(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, false)
	if err := r.Render(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sb.Len() != 0 {
		t.Errorf("expected empty output for nil report, got %q", sb.String())
	}
}

func TestRender_MatchEvent(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, false)
	report := &drift.Report{
		Events: []drift.Event{makeEvent("svc", "/etc/app.conf", drift.StatusMatch, "", "")},
	}
	if err := r.Render(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sb.String(), "(no drift)") {
		t.Errorf("expected '(no drift)' in output, got: %q", sb.String())
	}
}

func TestRender_MissingEvent(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, false)
	report := &drift.Report{
		Events: []drift.Event{makeEvent("svc", "/etc/app.conf", drift.StatusMissing, "", "")},
	}
	if err := r.Render(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "missing from live system") {
		t.Errorf("expected missing message, got: %q", out)
	}
	if !strings.HasPrefix(strings.Split(out, "\n")[1], "- ") {
		t.Errorf("expected removal prefix '-', got: %q", out)
	}
}

func TestRender_DriftedEvent(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, false)
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/app.conf", drift.StatusDrifted, "expected-value", "actual-value"),
		},
	}
	if err := r.Render(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "- expected-value") {
		t.Errorf("expected removal line, got: %q", out)
	}
	if !strings.Contains(out, "+ actual-value") {
		t.Errorf("expected addition line, got: %q", out)
	}
}

func TestRender_ColoursEnabled(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, true)
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/app.conf", drift.StatusMissing, "", ""),
		},
	}
	if err := r.Render(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(sb.String(), "\033[") {
		t.Errorf("expected ANSI escape codes in coloured output")
	}
}

func TestRender_MultilineContent(t *testing.T) {
	var sb strings.Builder
	r := diffrender.New(&sb, false)
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/hosts", drift.StatusDrifted, "line1\nline2", "lineA\nlineB"),
		},
	}
	if err := r.Render(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	for _, want := range []string{"- line1", "- line2", "+ lineA", "+ lineB"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %q", want, out)
		}
	}
}
