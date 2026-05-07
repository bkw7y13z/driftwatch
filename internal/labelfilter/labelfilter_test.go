package labelfilter_test

import (
	"sort"
	"testing"

	"github.com/yourorg/driftwatch/internal/labelfilter"
)

func TestNew_ValidExpressions(t *testing.T) {
	f, err := labelfilter.New([]string{"env=prod", "region=us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestNew_InvalidExpression(t *testing.T) {
	cases := []string{"noequalssign", "=nokey", ""}
	for _, expr := range cases {
		_, err := labelfilter.New([]string{expr})
		if err == nil {
			t.Errorf("expected error for expression %q, got nil", expr)
		}
	}
}

func TestMatches_AllSelectorsPresent(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=prod", "team=platform"})
	labels := labelfilter.Labels{"env": "prod", "team": "platform", "region": "eu"}
	if !f.Matches(labels) {
		t.Error("expected match")
	}
}

func TestMatches_MissingLabel(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=prod"})
	labels := labelfilter.Labels{"team": "platform"}
	if f.Matches(labels) {
		t.Error("expected no match when label key absent")
	}
}

func TestMatches_WrongValue(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=prod"})
	labels := labelfilter.Labels{"env": "staging"}
	if f.Matches(labels) {
		t.Error("expected no match when value differs")
	}
}

func TestMatches_EmptyFilter_MatchesAll(t *testing.T) {
	f, _ := labelfilter.New(nil)
	if !f.Matches(labelfilter.Labels{}) {
		t.Error("empty filter should match everything")
	}
}

func TestMatchAll_ReturnsSubset(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=prod"})
	services := map[string]labelfilter.Labels{
		"svc-a": {"env": "prod"},
		"svc-b": {"env": "staging"},
		"svc-c": {"env": "prod", "team": "ops"},
	}
	got := f.MatchAll(services)
	sort.Strings(got)
	if len(got) != 2 || got[0] != "svc-a" || got[1] != "svc-c" {
		t.Errorf("unexpected matched services: %v", got)
	}
}

func TestMatchAll_NoMatches(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=canary"})
	services := map[string]labelfilter.Labels{
		"svc-a": {"env": "prod"},
	}
	if got := f.MatchAll(services); len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestString_EmptyFilter(t *testing.T) {
	f, _ := labelfilter.New(nil)
	if f.String() != "<match-all>" {
		t.Errorf("unexpected string: %s", f.String())
	}
}

func TestString_WithSelectors(t *testing.T) {
	f, _ := labelfilter.New([]string{"env=prod"})
	if f.String() != "env=prod" {
		t.Errorf("unexpected string: %s", f.String())
	}
}
