package baseline_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/driftwatch/internal/baseline"
)

func newStore(t *testing.T) *baseline.Store {
	t.Helper()
	s, err := baseline.New(tempPath(t))
	if err != nil {
		t.Fatalf("baseline.New: %v", err)
	}
	return s
}

func TestHTTP_GetMissing_Returns404(t *testing.T) {
	h := baseline.Handler(newStore(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/baseline/api", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

func TestHTTP_PostAndGet_RoundTrip(t *testing.T) {
	h := baseline.Handler(newStore(t))

	body, _ := json.Marshal(baseline.Entry{
		Ref:   "sha999",
		Files: map[string]string{"svc.yaml": "aabbcc"},
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/baseline/mysvc", bytes.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST: want 201, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/baseline/mysvc", nil))
	if rec2.Code != http.StatusOK {
		t.Fatalf("GET: want 200, got %d", rec2.Code)
	}
	var got baseline.Entry
	_ = json.NewDecoder(rec2.Body).Decode(&got)
	if got.Ref != "sha999" {
		t.Errorf("ref: want sha999, got %s", got.Ref)
	}
	if got.Service != "mysvc" {
		t.Errorf("service overridden from URL: want mysvc, got %s", got.Service)
	}
}

func TestHTTP_Delete_Removes(t *testing.T) {
	s := newStore(t)
	_ = s.Pin(baseline.Entry{Service: "del-svc", Ref: "r", Files: map[string]string{}})
	h := baseline.Handler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/baseline/del-svc", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("DELETE: want 204, got %d", rec.Code)
	}
	_, ok := s.Get("del-svc")
	if ok {
		t.Error("entry should be gone after DELETE")
	}
}

func TestHTTP_MissingServiceName_Returns400(t *testing.T) {
	h := baseline.Handler(newStore(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/baseline/", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", rec.Code)
	}
}

func TestHTTP_MethodNotAllowed(t *testing.T) {
	h := baseline.Handler(newStore(t))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/baseline/svc", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", rec.Code)
	}
}
