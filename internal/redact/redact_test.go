package redact

import (
	"strings"
	"testing"
)

func TestNew_DefaultPatterns(t *testing.T) {
	r, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.patterns) != len(DefaultPatterns) {
		t.Fatalf("expected %d patterns, got %d", len(DefaultPatterns), len(r.patterns))
	}
}

func TestNew_InvalidPattern(t *testing.T) {
	_, err := New([]string{"["})
	if err == nil {
		t.Fatal("expected error for invalid pattern")
	}
}

func TestIsSensitiveKey(t *testing.T) {
	r, _ := New(nil)
	cases := []struct {
		key       string
		wantMatch bool
	}{
		{"password", true},
		{"db_password", true},
		{"PASSWORD", true},
		{"api_key", true},
		{"apikey", true},
		{"token", true},
		{"secret", true},
		{"host", false},
		{"port", false},
		{"name", false},
	}
	for _, tc := range cases {
		got := r.IsSensitiveKey(tc.key)
		if got != tc.wantMatch {
			t.Errorf("IsSensitiveKey(%q) = %v, want %v", tc.key, got, tc.wantMatch)
		}
	}
}

func TestApply_RedactsMatchingKeys(t *testing.T) {
	r, _ := New(nil)
	input := "host: localhost\npassword: s3cr3t\nport: 5432\n"
	out := r.Apply(input)
	if strings.Contains(out, "s3cr3t") {
		t.Errorf("sensitive value not redacted; got:\n%s", out)
	}
	if !strings.Contains(out, redactedPlaceholder) {
		t.Errorf("expected placeholder in output; got:\n%s", out)
	}
	if !strings.Contains(out, "host: localhost") {
		t.Errorf("non-sensitive key should be unchanged; got:\n%s", out)
	}
}

func TestApply_PreservesIndentation(t *testing.T) {
	r, _ := New(nil)
	input := "  token: abc123"
	out := r.Apply(input)
	if !strings.HasPrefix(out, "  ") {
		t.Errorf("expected leading whitespace preserved; got: %q", out)
	}
	if strings.Contains(out, "abc123") {
		t.Errorf("value should be redacted; got: %q", out)
	}
}

func TestApply_NoColonLine(t *testing.T) {
	r, _ := New(nil)
	input := "just a plain line"
	out := r.Apply(input)
	if out != input {
		t.Errorf("expected unchanged line; got: %q", out)
	}
}

func TestApply_CustomPatterns(t *testing.T) {
	r, err := New([]string{"internal_id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	input := "internal_id: 42\npassword: should_stay"
	out := r.Apply(input)
	if !strings.Contains(out, redactedPlaceholder) {
		t.Error("expected custom key to be redacted")
	}
	// default patterns should NOT apply with custom patterns
	if !strings.Contains(out, "should_stay") {
		t.Error("non-custom key should not be redacted")
	}
}
