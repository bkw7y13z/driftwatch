package baseline_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/driftwatch/internal/baseline"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "baseline.json")
}

func TestPin_And_Get(t *testing.T) {
	s, err := baseline.New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	e := baseline.Entry{
		Service: "api",
		Ref:     "abc123",
		Files:   map[string]string{"config.yaml": "deadbeef"},
	}
	if err := s.Pin(e); err != nil {
		t.Fatalf("Pin: %v", err)
	}
	got, ok := s.Get("api")
	if !ok {
		t.Fatal("expected entry, got none")
	}
	if got.Ref != "abc123" {
		t.Errorf("ref: want abc123, got %s", got.Ref)
	}
	if got.PinnedAt.IsZero() {
		t.Error("PinnedAt should be set")
	}
}

func TestGet_Missing(t *testing.T) {
	s, _ := baseline.New(tempPath(t))
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected false for missing service")
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	s, _ := baseline.New(tempPath(t))
	_ = s.Pin(baseline.Entry{Service: "svc", Ref: "r1", Files: map[string]string{}})
	if err := s.Remove("svc"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := s.Get("svc")
	if ok {
		t.Error("entry should be gone after Remove")
	}
}

func TestPersistence_ReloadFromDisk(t *testing.T) {
	p := tempPath(t)
	s1, _ := baseline.New(p)
	_ = s1.Pin(baseline.Entry{Service: "worker", Ref: "v2", Files: map[string]string{"app.conf": "cafebabe"}})

	s2, err := baseline.New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	got, ok := s2.Get("worker")
	if !ok {
		t.Fatal("entry missing after reload")
	}
	if got.Files["app.conf"] != "cafebabe" {
		t.Errorf("file hash mismatch after reload")
	}
}

func TestNew_MissingFile_IsOK(t *testing.T) {
	p := filepath.Join(t.TempDir(), "does-not-exist.json")
	s, err := baseline.New(p)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestNew_CorruptFile_ReturnsError(t *testing.T) {
	p := tempPath(t)
	_ = os.WriteFile(p, []byte("not json{"), 0o644)
	_, err := baseline.New(p)
	if err == nil {
		t.Error("expected error for corrupt file")
	}
}
