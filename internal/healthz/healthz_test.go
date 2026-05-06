package healthz_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/driftwatch/internal/healthz"
)

func TestServeHTTP_BeforeAnyCheck(t *testing.T) {
	h := healthz.New(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "starting" {
		t.Errorf("expected status=starting, got %v", body["status"])
	}
}

func TestServeHTTP_AfterCheck_NoDrift(t *testing.T) {
	h := healthz.New(nil)
	h.RecordCheck(false)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
	if body["drifted_seen"].(bool) {
		t.Error("expected drifted_seen=false")
	}
	if body["last_check"] == "" {
		t.Error("expected last_check to be set")
	}
}

func TestServeHTTP_AfterCheck_WithDrift(t *testing.T) {
	h := healthz.New(nil)
	h.RecordCheck(true)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !body["drifted_seen"].(bool) {
		t.Error("expected drifted_seen=true")
	}
}

func TestRegister_MountsAtHealthz(t *testing.T) {
	h := healthz.New(nil)
	h.RecordCheck(false)

	mux := http.NewServeMux()
	healthz.Register(mux, h)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 via mux, got %d", rec.Code)
	}
}
