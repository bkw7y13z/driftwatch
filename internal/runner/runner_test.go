package runner

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/config"
)

func makeTestConfig(t *testing.T, repoPath string) *config.Config {
	t.Helper()
	return &config.Config{
		RepoPath:        repoPath,
		GitRef:          "HEAD",
		IntervalSeconds: 60,
		Services: []config.Service{
			{Name: "svc-a", ConfigPath: "configs/svc-a.yaml"},
		},
	}
}

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestNew_InvalidRepo(t *testing.T) {
	cfg := makeTestConfig(t, "/nonexistent/path")
	_, err := New(cfg, silentLogger())
	if err == nil {
		t.Fatal("expected error for invalid repo path, got nil")
	}
}

func TestNew_ValidRepo(t *testing.T) {
	repo := t.TempDir()
	// initialise a bare git repo so NewFetcher is satisfied
	if err := initBareRepo(t, repo); err != nil {
		t.Skipf("git not available: %v", err)
	}

	cfg := makeTestConfig(t, repo)
	r, err := New(cfg, silentLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
}

func TestStart_CancelImmediately(t *testing.T) {
	repo := t.TempDir()
	if err := initBareRepo(t, repo); err != nil {
		t.Skipf("git not available: %v", err)
	}

	cfg := makeTestConfig(t, repo)
	cfg.IntervalSeconds = 5

	r, err := New(cfg, silentLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = r.Start(ctx)
	if err != context.DeadlineExceeded && err != context.Canceled {
		t.Fatalf("expected context error, got: %v", err)
	}
}

// initBareRepo creates a minimal git repo in dir using the git CLI.
func initBareRepo(t *testing.T, dir string) error {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0o755)
	// Write a minimal HEAD so RepoExists passes.
	return os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644)
}
