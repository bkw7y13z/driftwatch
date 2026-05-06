package watcher_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/driftwatch/internal/watcher"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	return p
}

func TestReadFile_Found(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app.conf", "key=value\n")

	w := watcher.New([]string{dir}, nil)
	state := w.ReadFile(context.Background(), "app.conf")

	if state.Missing {
		t.Fatal("expected file to be found")
	}
	if string(state.Content) != "key=value\n" {
		t.Fatalf("unexpected content: %q", state.Content)
	}
}

func TestReadFile_Missing(t *testing.T) {
	dir := t.TempDir()
	w := watcher.New([]string{dir}, nil)
	state := w.ReadFile(context.Background(), "ghost.conf")

	if !state.Missing {
		t.Fatal("expected file to be missing")
	}
}

func TestReadFile_FallsBackToSecondBase(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	writeFile(t, dir2, "svc.yaml", "name: svc\n")

	w := watcher.New([]string{dir1, dir2}, nil)
	state := w.ReadFile(context.Background(), "svc.yaml")

	if state.Missing {
		t.Fatal("expected file found in second base path")
	}
}

func TestReadAll_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	w := watcher.New([]string{dir}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := w.ReadAll(ctx, []string{"a.conf", "b.conf"})
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}

func TestToLiveContent_ExcludesMissing(t *testing.T) {
	states := map[string]watcher.FileState{
		"present.conf": {Path: "/etc/present.conf", Content: []byte("ok")},
		"absent.conf":  {Path: "absent.conf", Missing: true},
	}

	live := watcher.ToLiveContent(states)
	if _, ok := live["absent.conf"]; ok {
		t.Error("missing file should not appear in live content")
	}
	if _, ok := live["present.conf"]; !ok {
		t.Error("present file should appear in live content")
	}
}

func TestMissingPaths(t *testing.T) {
	states := map[string]watcher.FileState{
		"a.conf": {Missing: false},
		"b.conf": {Missing: true},
		"c.conf": {Missing: true},
	}
	missing := watcher.MissingPaths(states)
	if len(missing) != 2 {
		t.Fatalf("expected 2 missing paths, got %d", len(missing))
	}
}
