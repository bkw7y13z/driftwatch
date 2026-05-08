package policycheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/policycheck"
)

type stubStore struct{ report *drift.Report }

func (s *stubStore) Latest() *drift.Report { return s.report }

func TestHTTP_NoViolations_Returns204(t *testing.T) {
	c, _ := policycheck.New([]policycheck.Policy{
		{Name: "no-drift", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityWarn},
	})
	store := &stubStore{report: makeReport(drift.Event{Service: "api", Status: drift.StatusMatch})}

	rec := httptest.NewRecorder()
	policycheck.Handler(c, store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHTTP_WithViolations_Returns409(t *testing.T) {
	c, _ := policycheck.New([]policycheck.Policy{
		{Name: "no-drift", ServicePattern: ".*", DenyStatuses: []drift.Status{drift.StatusDrifted}, Severity: policycheck.SeverityError},
	})
	store := &stubStore{report: makeReport(drift.Event{Service: "api", Status: drift.StatusDrifted})}

	rec := httptest.NewRecorder()
	policycheck.Handler(c, store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}

	var violations []policycheck.Violation
	if err := json.NewDecoder(rec.Body).Decode(&violations); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation in body, got %d", len(violations))
	}
	if violations[0].Policy != "no-drift" {
		t.Errorf("unexpected policy name: %s", violations[0].Policy)
	}
}

func TestHTTP_NilReport_Returns204(t *testing.T) {
	c, _ := policycheck.New(nil)
	store := &stubStore{report: nil}

	rec := httptest.NewRecorder()
	policycheck.Handler(c, store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}
