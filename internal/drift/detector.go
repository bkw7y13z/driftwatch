package drift

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Status represents the drift state of a single file.
type Status int

const (
	StatusMatch   Status = iota // file content matches declared state
	StatusDrifted               // file content differs from declared state
	StatusMissing               // file is missing from the running service
)

// Result holds the drift detection outcome for one file.
type Result struct {
	FilePath  string
	Status    Status
	Expected  string // sha256 of declared content
	Actual    string // sha256 of live content (empty if missing)
}

// String returns a human-readable summary of the result.
func (r Result) String() string {
	switch r.Status {
	case StatusMatch:
		return fmt.Sprintf("[OK]      %s", r.FilePath)
	case StatusDrifted:
		return fmt.Sprintf("[DRIFT]   %s (expected %s, got %s)", r.FilePath, r.Expected[:8], r.Actual[:8])
	case StatusMissing:
		return fmt.Sprintf("[MISSING] %s", r.FilePath)
	default:
		return fmt.Sprintf("[UNKNOWN] %s", r.FilePath)
	}
}

// Detector compares declared file contents against live contents.
type Detector struct{}

// NewDetector creates a new Detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Compare checks whether liveContent matches declaredContent for the given path.
// Pass an empty liveContent to indicate the file is missing.
func (d *Detector) Compare(filePath, declaredContent, liveContent string) Result {
	expectedHash := hashContent(declaredContent)

	if liveContent == "" {
		return Result{
			FilePath: filePath,
			Status:   StatusMissing,
			Expected: expectedHash,
			Actual:   "",
		}
	}

	actualHash := hashContent(liveContent)
	status := StatusMatch
	if expectedHash != actualHash {
		status = StatusDrifted
	}

	return Result{
		FilePath: filePath,
		Status:   status,
		Expected: expectedHash,
		Actual:   actualHash,
	}
}

// CompareAll runs Compare across a map of filePath -> declared content,
// using the provided live content map.
func (d *Detector) CompareAll(declared, live map[string]string) []Result {
	results := make([]Result, 0, len(declared))
	for path, declaredContent := range declared {
		liveContent := live[path]
		results = append(results, d.Compare(path, declaredContent, liveContent))
	}
	return results
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(strings.TrimRight(content, "\n")))
	return fmt.Sprintf("%x", h)
}
