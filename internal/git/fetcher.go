package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Fetcher retrieves file contents from a Git repository at a specific ref.
type Fetcher struct {
	repoPath string
	repo     *git.Repository
}

// NewFetcher opens a local git repository at the given path.
func NewFetcher(repoPath string) (*Fetcher, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("resolving repo path: %w", err)
	}

	repo, err := git.PlainOpen(absPath)
	if err != nil {
		return nil, fmt.Errorf("opening git repo at %s: %w", absPath, err)
	}

	return &Fetcher{repoPath: absPath, repo: repo}, nil
}

// ReadFileAtRef returns the contents of a file at the given git ref (branch, tag, or commit SHA).
func (f *Fetcher) ReadFileAtRef(ref, filePath string) ([]byte, error) {
	hash, err := f.resolveRef(ref)
	if err != nil {
		return nil, fmt.Errorf("resolving ref %q: %w", ref, err)
	}

	commit, err := f.repo.CommitObject(hash)
	if err != nil {
		return nil, fmt.Errorf("fetching commit %s: %w", hash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("fetching tree for commit %s: %w", hash, err)
	}

	entry, err := tree.File(filePath)
	if err != nil {
		return nil, fmt.Errorf("file %q not found at ref %q: %w", filePath, ref, err)
	}

	contents, err := entry.Contents()
	if err != nil {
		return nil, fmt.Errorf("reading file %q: %w", filePath, err)
	}

	return []byte(contents), nil
}

// HeadRef returns the short name of the current HEAD branch.
func (f *Fetcher) HeadRef() (string, error) {
	head, err := f.repo.Head()
	if err != nil {
		return "", fmt.Errorf("reading HEAD: %w", err)
	}
	return head.Name().Short(), nil
}

func (f *Fetcher) resolveRef(ref string) (plumbing.Hash, error) {
	// Try as a branch or tag first.
	hash, err := f.repo.ResolveRevision(plumbing.Revision(ref))
	if err == nil {
		return *hash, nil
	}
	// Fall back to treating it as a commit SHA.
	h := plumbing.NewHash(ref)
	if h == plumbing.ZeroHash {
		return plumbing.ZeroHash, fmt.Errorf("invalid ref or hash: %q", ref)
	}
	return h, nil
}

// RepoExists returns true if the path contains a valid git repository.
func RepoExists(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
