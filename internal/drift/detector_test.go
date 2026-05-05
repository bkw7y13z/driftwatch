package drift

import (
	"strings"
	"testing"
)

func TestCompare_Match(t *testing.T) {
	d := NewDetector()
	content := "key: value\nother: 123\n"
	result := d.Compare("config.yaml", content, content)

	if result.Status != StatusMatch {
		t.Errorf("expected StatusMatch, got %v", result.Status)
	}
	if result.FilePath != "config.yaml" {
		t.Errorf("unexpected filepath: %s", result.FilePath)
	}
}

func TestCompare_Drifted(t *testing.T) {
	d := NewDetector()
	declared := "key: value\n"
	live := "key: changed\n"
	result := d.Compare("app.conf", declared, live)

	if result.Status != StatusDrifted {
		t.Errorf("expected StatusDrifted, got %v", result.Status)
	}
	if result.Expected == result.Actual {
		t.Error("expected different hashes for drifted content")
	}
}

func TestCompare_Missing(t *testing.T) {
	d := NewDetector()
	result := d.Compare("missing.yaml", "declared content", "")

	if result.Status != StatusMissing {
		t.Errorf("expected StatusMissing, got %v", result.Status)
	}
	if result.Actual != "" {
		t.Errorf("expected empty actual hash, got %s", result.Actual)
	}
}

func TestCompare_TrailingNewlineIgnored(t *testing.T) {
	d := NewDetector()
	withNewline := "key: value\n"
	withoutNewline := "key: value"
	result := d.Compare("cfg.yaml", withNewline, withoutNewline)

	if result.Status != StatusMatch {
		t.Errorf("trailing newline difference should not cause drift, got %v", result.Status)
	}
}

func TestCompareAll(t *testing.T) {
	d := NewDetector()
	declared := map[string]string{
		"a.yaml": "a: 1",
		"b.yaml": "b: 2",
		"c.yaml": "c: 3",
	}
	live := map[string]string{
		"a.yaml": "a: 1",  // match
		"b.yaml": "b: 99", // drift
		// c.yaml missing
	}

	results := d.CompareAll(declared, live)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	counts := map[Status]int{}
	for _, r := range results {
		counts[r.Status]++
	}
	if counts[StatusMatch] != 1 || counts[StatusDrifted] != 1 || counts[StatusMissing] != 1 {
		t.Errorf("unexpected status counts: %v", counts)
	}
}

func TestResult_String(t *testing.T) {
	d := NewDetector()
	r := d.Compare("srv.yaml", "x: 1", "x: 2")
	s := r.String()
	if !strings.HasPrefix(s, "[DRIFT]") {
		t.Errorf("expected [DRIFT] prefix, got: %s", s)
	}
}
