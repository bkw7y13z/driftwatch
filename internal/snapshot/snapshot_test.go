package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/snapshot"
)

func TestNew(t *testing.T) {
	s := snapshot.New("svc-a")
	if s.ServiceName != "svc-a" {
		t.Fatalf("expected service name svc-a, got %s", s.ServiceName)
	}
	if s.Entries == nil {
		t.Fatal("expected non-nil entries map")
	}
	if s.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestSetAndGet(t *testing.T) {
	s := snapshot.New("svc-b")
	s.Set("config/app.yaml", "abc123", "main")

	e, ok := s.Get("config/app.yaml")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if e.Hash != "abc123" {
		t.Errorf("expected hash abc123, got %s", e.Hash)
	}
	if e.Ref != "main" {
		t.Errorf("expected ref main, got %s", e.Ref)
	}
	if e.RecordedAt.IsZero() {
		t.Error("expected non-zero RecordedAt")
	}
}

func TestGet_Missing(t *testing.T) {
	s := snapshot.New("svc-c")
	_, ok := s.Get("nonexistent.yaml")
	if ok {
		t.Fatal("expected false for missing entry")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	s := snapshot.New("svc-save")
	s.Set("deploy/k8s.yaml", "deadbeef", "v1.2.3")

	if err := s.Save(dir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	expected := filepath.Join(dir, "svc-save.snapshot.json")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected snapshot file at %s: %v", expected, err)
	}

	loaded, err := snapshot.Load(dir, "svc-save")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if loaded.ServiceName != "svc-save" {
		t.Errorf("expected svc-save, got %s", loaded.ServiceName)
	}
	e, ok := loaded.Get("deploy/k8s.yaml")
	if !ok {
		t.Fatal("expected entry in loaded snapshot")
	}
	if e.Hash != "deadbeef" {
		t.Errorf("expected hash deadbeef, got %s", e.Hash)
	}
}

func TestLoad_NoFile(t *testing.T) {
	dir := t.TempDir()
	snap, err := snapshot.Load(dir, "nonexistent-svc")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if snap != nil {
		t.Fatal("expected nil snapshot when file does not exist")
	}
}

func TestSet_Overwrite(t *testing.T) {
	s := snapshot.New("svc-ow")
	s.Set("file.yaml", "hash1", "ref1")
	time.Sleep(time.Millisecond)
	s.Set("file.yaml", "hash2", "ref2")

	e, _ := s.Get("file.yaml")
	if e.Hash != "hash2" {
		t.Errorf("expected overwritten hash hash2, got %s", e.Hash)
	}
}
