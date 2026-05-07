package fingerprint_test

import (
	"strings"
	"testing"

	"github.com/driftwatch/driftwatch/internal/fingerprint"
)

func TestSum_ProducesConsistentResult(t *testing.T) {
	a := fingerprint.Sum("hello world")
	b := fingerprint.Sum("hello world")
	if !a.Equal(b) {
		t.Fatalf("expected equal fingerprints, got %s and %s", a, b)
	}
}

func TestSum_DifferentContentDiffers(t *testing.T) {
	a := fingerprint.Sum("foo")
	b := fingerprint.Sum("bar")
	if a.Equal(b) {
		t.Fatal("expected different fingerprints for different content")
	}
}

func TestSum_NormalisesWhitespace(t *testing.T) {
	a := fingerprint.Sum("config: value")
	b := fingerprint.Sum("  config: value  ")
	if !a.Equal(b) {
		t.Fatalf("expected equal after trim, got %s and %s", a, b)
	}
}

func TestResult_String_Format(t *testing.T) {
	r := fingerprint.Sum("test")
	s := r.String()
	if !strings.HasPrefix(s, "sha256:") {
		t.Fatalf("expected sha256: prefix, got %q", s)
	}
	if len(s) != len("sha256:")+64 {
		t.Fatalf("unexpected length %d for %q", len(s), s)
	}
}

func TestParse_RoundTrip(t *testing.T) {
	orig := fingerprint.Sum("round trip content")
	parsed, ok := fingerprint.Parse(orig.String())
	if !ok {
		t.Fatal("Parse returned ok=false for valid fingerprint")
	}
	if !orig.Equal(parsed) {
		t.Fatalf("round-trip mismatch: %s != %s", orig, parsed)
	}
}

func TestParse_InvalidFormat(t *testing.T) {
	cases := []string{"", "nocolon", "unknown:abc123"}
	for _, c := range cases {
		_, ok := fingerprint.Parse(c)
		if ok {
			t.Errorf("expected Parse(%q) to fail", c)
		}
	}
}

func TestParse_UnknownAlgorithm(t *testing.T) {
	_, ok := fingerprint.Parse("md5:d41d8cd98f00b204e9800998ecf8427e")
	if ok {
		t.Fatal("expected Parse to reject unknown algorithm")
	}
}
