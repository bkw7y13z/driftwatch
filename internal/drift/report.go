package drift

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// Report aggregates drift results for a named service at a point in time.
type Report struct {
	ServiceName string
	Ref         string
	CheckedAt   time.Time
	Results     []Result
}

// HasDrift returns true if any result is not StatusMatch.
func (r *Report) HasDrift() bool {
	for _, res := range r.Results {
		if res.Status != StatusMatch {
			return true
		}
	}
	return false
}

// Summary returns counts of each status.
func (r *Report) Summary() (matched, drifted, missing int) {
	for _, res := range r.Results {
		switch res.Status {
		case StatusMatch:
			matched++
		case StatusDrifted:
			drifted++
		case StatusMissing:
			missing++
		}
	}
	return
}

// WriteTo writes a human-readable report to w.
func (r *Report) WriteTo(w io.Writer) error {
	sep := strings.Repeat("-", 60)
	_, err := fmt.Fprintf(w,
		"%s\nDrift Report: %s @ %s\nChecked: %s\n%s\n",
		sep, r.ServiceName, r.Ref,
		r.CheckedAt.Format(time.RFC3339),
		sep,
	)
	if err != nil {
		return err
	}

	for _, res := range r.Results {
		if _, err := fmt.Fprintln(w, res.String()); err != nil {
			return err
		}
	}

	matched, drifted, missing := r.Summary()
	_, err = fmt.Fprintf(w,
		"%s\nSummary: %d matched, %d drifted, %d missing\n",
		sep, matched, drifted, missing,
	)
	return err
}
