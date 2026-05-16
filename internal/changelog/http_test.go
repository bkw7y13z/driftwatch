package changelog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP_Get_EmptyLog(t *testing.T) {
	l := New(10)
	rec := httptest.NewRecorder()
	Handler(l).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/changelog", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out []entryJSON
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty array, got %d items", len(out))
	}
}

func TestHTTP_Get_ReturnsEntries(t *testing.T) {
	l := New(10)
	l.Record(makeReport("svc-a", "svc-b"))
	rec := httptest.NewRecorder()
	Handler(l).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/changelog", nil))
	var out []entryJSON
	_ = json.NewDecoder(rec.Body).Decode(&out)
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
}

func TestHTTP_Delete_ClearsLog(t *testing.T) {
	l := New(10)
	l.Record(makeReport("svc-a"))
	rec := httptest.NewRecorder()
	Handler(l).ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/changelog", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := len(l.Entries()); got != 0 {
		t.Fatalf("expected 0 entries after delete, got %d", got)
	}
}

func TestHTTP_MethodNotAllowed(t *testing.T) {
	l := New(10)
	rec := httptest.NewRecorder()
	Handler(l).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/changelog", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
