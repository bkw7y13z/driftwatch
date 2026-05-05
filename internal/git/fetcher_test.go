package git_test

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/yourorg/driftwatch/internal/git"
)

// initTestRepo creates a temporary git repo with a single committed file.
func initTestRepo(t *testing.T, fileName, content string) string {
	t.Helper()
	dir := t.TempDir()

	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}

	filePath := filepath.Join(dir, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("worktree: %v", err)
	}
	if _, err := wt.Add(fileName); err != nil {
		t.Fatalf("git add: %v", err)
	}
	_, err = wt.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@example.com"},
	})
	if err != nil {
		t.Fatalf("git commit: %v", err)
	}

	return dir
}

func TestNewFetcher_ValidRepo(t *testing.T) {
	dir := initTestRepo(t, "service.yaml", "name: svc")
	_, err := git.NewFetcher(dir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestNewFetcher_InvalidPath(t *testing.T) {
	_, err := git.NewFetcher("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

func TestReadFileAtRef(t *testing.T) {
	const fileContent = "replicas: 3\nimage: myapp:latest\n"
	dir := initTestRepo(t, "deploy.yaml", fileContent)

	f, err := git.NewFetcher(dir)
	if err != nil {
		t.Fatalf("NewFetcher: %v", err)
	}

	ref, err := f.HeadRef()
	if err != nil {
		t.Fatalf("HeadRef: %v", err)
	}

	got, err := f.ReadFileAtRef(ref, "deploy.yaml")
	if err != nil {
		t.Fatalf("ReadFileAtRef: %v", err)
	}
	if string(got) != fileContent {
		t.Errorf("content mismatch: got %q, want %q", got, fileContent)
	}
}

func TestReadFileAtRef_MissingFile(t *testing.T) {
	dir := initTestRepo(t, "exists.yaml", "key: val")
	f, err := git.NewFetcher(dir)
	if err != nil {
		t.Fatalf("NewFetcher: %v", err)
	}
	ref, _ := f.HeadRef()
	_, err = f.ReadFileAtRef(ref, "missing.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestRepoExists(t *testing.T) {
	dir := initTestRepo(t, "f.txt", "x")
	if !git.RepoExists(dir) {
		t.Errorf("expected RepoExists=true for %s", dir)
	}
	if git.RepoExists(t.TempDir()) {
		t.Error("expected RepoExists=false for plain dir")
	}
}
