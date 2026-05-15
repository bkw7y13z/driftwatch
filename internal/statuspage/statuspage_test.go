package statuspage_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/statuspage"
)

func makeReport(service string, events []drift.Event) *drift.Report {
	return &drift.Report{Service: service, Events: events}
}

func TestRecord_NilReport(t *testing.T) {
	p := statuspage.New()
	p.Record(nil) // must not panic
	if got := p.Statuses(); len(got) != 0 {
		t.Fatalf("expected 0 statuses, got %d", len(got))
	}
}

func TestRecord_StoresStatus(t *testing.T) {
	p := statuspage.New()
	p.Record(makeReport("svc-a", []drift.Event{
		{Path: "config.yaml", Status: drift.StatusMatch},
	}))

	statuses := p.Statuses()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Service != "svc-a" {
		t.Errorf("unexpected service name: %s", statuses[0].Service)
	}
	if statuses[0].HasDrift {
		t.Error("expected no drift for all-match report")
	}
}

func TestRecord_PreservesDriftedAt(t *testing.T) {
	p := statuspage.New()
	driftedEvent := []drift.Event{{Path: "app.conf", Status: drift.StatusDrifted}}

	p.Record(makeReport("svc-b", driftedEvent))
	first := p.Statuses()
	if len(first) != 1 || !first[0].HasDrift {
		t.Fatal("expected drifted status after first record")
	}
	firstDriftedAt := first[0].DriftedAt

	p.Record(makeReport("svc-b", driftedEvent))
	second := p.Statuses()
	if second[0].DriftedAt != firstDriftedAt {
		t.Error("DriftedAt should be preserved across consecutive drifted records")
	}
}

func TestServeHTTP_NoDrift_Returns200(t *testing.T) {
	p := statuspage.New()
	p.Record(makeReport("svc-ok", []drift.Event{
		{Path: "config.yaml", Status: drift.StatusMatch},
	}))

	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statuspage", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_WithDrift_Returns409(t *testing.T) {
	p := statuspage.New()
	p.Record(makeReport("svc-bad", []drift.Event{
		{Path: "config.yaml", Status: drift.StatusDrifted},
	}))

	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statuspage", nil))

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestServeHTTP_ResponseIsValidJSON(t *testing.T) {
	p := statuspage.New()
	p.Record(makeReport("svc-json", []drift.Event{
		{Path: "app.yaml", Status: drift.StatusMatch},
	}))

	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statuspage", nil))

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := body["services"]; !ok {
		t.Error("response JSON missing 'services' key")
	}
	if _, ok := body["generated_at"]; !ok {
		t.Error("response JSON missing 'generated_at' key")
	}
}

func TestRegister_MountsHandler(t *testing.T) {
	p := statuspage.New()
	mux := http.NewServeMux()
	statuspage.Register(mux, p)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/statuspage", nil))

	if rec.Code == http.StatusNotFound {
		t.Error("handler not mounted: got 404")
	}
}
