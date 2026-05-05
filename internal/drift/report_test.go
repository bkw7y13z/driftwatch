package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeReport(results []Result) *Report {
	return &Report{
		ServiceName: "my-service",
		Ref:         "main",
		CheckedAt:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Results:     results,
	}
}

func TestReport_HasDrift_False(t *testing.T) {
	r := makeReport([]Result{
		{FilePath: "a.yaml", Status: StatusMatch},
		{FilePath: "b.yaml", Status: StatusMatch},
	})
	if r.HasDrift() {
		t.Error("expected no drift")
	}
}

func TestReport_HasDrift_True(t *testing.T) {
	r := makeReport([]Result{
		{FilePath: "a.yaml", Status: StatusMatch},
		{FilePath: "b.yaml", Status: StatusDrifted},
	})
	if !r.HasDrift() {
		t.Error("expected drift to be detected")
	}
}

func TestReport_Summary(t *testing.T) {
	r := makeReport([]Result{
		{Status: StatusMatch},
		{Status: StatusMatch},
		{Status: StatusDrifted},
		{Status: StatusMissing},
	})
	matched, drifted, missing := r.Summary()
	if matched != 2 || drifted != 1 || missing != 1 {
		t.Errorf("unexpected summary: %d/%d/%d", matched, drifted, missing)
	}
}

func TestReport_WriteTo(t *testing.T) {
	d := NewDetector()
	results := []Result{
		d.Compare("cfg.yaml", "x: 1", "x: 1"),
		d.Compare("svc.yaml", "y: 2", "y: 9"),
		d.Compare("sec.yaml", "z: 3", ""),
	}
	r := makeReport(results)

	var buf bytes.Buffer
	if err := r.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"my-service",
		"main",
		"[OK]",
		"[DRIFT]",
		"[MISSING]",
		"Summary:",
		"1 matched",
		"1 drifted",
		"1 missing",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}
