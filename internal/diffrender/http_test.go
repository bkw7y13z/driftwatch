package diffrender_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/diffrender"
	"github.com/example/driftwatch/internal/drift"
)

func TestHTTPHandler_NilReport_NoContent(t *testing.T) {
	h := diffrender.Handler(func() *drift.Report { return nil })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/diff", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHTTPHandler_NoDrift_Returns200(t *testing.T) {
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/app.conf", drift.StatusMatch, "", ""),
		},
	}
	h := diffrender.Handler(func() *drift.Report { return report })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/diff", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("unexpected content-type: %q", ct)
	}
}

func TestHTTPHandler_WithDrift_Returns409(t *testing.T) {
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/app.conf", drift.StatusDrifted, "want", "got"),
		},
	}
	h := diffrender.Handler(func() *drift.Report { return report })
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/diff", nil))
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "- want") || !strings.Contains(body, "+ got") {
		t.Errorf("expected diff lines in body, got: %q", body)
	}
}

func TestHTTPHandler_AnsiAccept_EnablesColours(t *testing.T) {
	report := &drift.Report{
		Events: []drift.Event{
			makeEvent("svc", "/etc/app.conf", drift.StatusMissing, "", ""),
		},
	}
	h := diffrender.Handler(func() *drift.Report { return report })
	req := httptest.NewRequest(http.MethodGet, "/diff", nil)
	req.Header.Set("Accept", "text/x-ansi")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "\033[") {
		t.Errorf("expected ANSI codes in response body")
	}
}
