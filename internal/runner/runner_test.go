package runner_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/config"
	"github.com/yourorg/driftwatch/internal/runner"
)

func silentLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func makeTestConfig(t *testing.T) *config.Config {
	t.Helper()
	repoDir := t.TempDir()
	// minimal bare git init so fetcher accepts it
	if err := os.MkdirAll(repoDir+"/.git", 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	return &config.Config{
		ServiceName:  "test-svc",
		RepoPath:     repoDir,
		GitRef:       "HEAD",
		Interval:     100 * time.Millisecond,
		WatchPaths:   []string{t.TempDir()},
		WatchFiles:   []string{},
		SnapshotPath: t.TempDir() + "/snap.json",
	}
}

func TestNew_InvalidRepo(t *testing.T) {
	cfg := &config.Config{
		ServiceName: "svc",
		RepoPath:    "/nonexistent/path/xyz",
		Interval:    time.Second,
	}
	_, err := runner.New(cfg, silentLogger())
	if err == nil {
		t.Fatal("expected error for invalid repo path")
	}
}

func TestNew_ValidRepo(t *testing.T) {
	cfg := makeTestConfig(t)
	_, err := runner.New(cfg, silentLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStart_CancelImmediately(t *testing.T) {
	cfg := makeTestConfig(t)
	r, err := runner.New(cfg, silentLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start returned unexpected error: %v", err)
	}
}

func TestStart_RunsAtLeastOnce(t *testing.T) {
	cfg := makeTestConfig(t)
	cfg.Interval = 50 * time.Millisecond
	r, err := runner.New(cfg, silentLogger())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Should complete without error even with empty watch files.
	if err := r.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
}
