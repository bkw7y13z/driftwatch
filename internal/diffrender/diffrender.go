// Package diffrender formats drift detection results into human-readable
// unified-diff-style output suitable for logs, terminals, and HTTP responses.
package diffrender

import (
	"fmt"
	"io"
	"strings"

	"github.com/example/driftwatch/internal/drift"
)

// Renderer writes formatted diff output for drift reports.
type Renderer struct {
	w       io.Writer
	colours bool
}

const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

// New returns a Renderer that writes to w. If colours is true, ANSI colour
// codes are included in the output.
func New(w io.Writer, colours bool) *Renderer {
	return &Renderer{w: w, colours: colours}
}

// Render writes a human-readable diff summary for every event in report.
func (r *Renderer) Render(report *drift.Report) error {
	if report == nil {
		return nil
	}
	for _, ev := range report.Events {
		if err := r.renderEvent(ev); err != nil {
			return err
		}
	}
	return nil
}

func (r *Renderer) renderEvent(ev drift.Event) error {
	header := fmt.Sprintf("--- %s [%s]\n", ev.Path, ev.Service)
	if _, err := fmt.Fprint(r.w, header); err != nil {
		return err
	}

	switch ev.Status {
	case drift.StatusMatch:
		_, err := fmt.Fprintf(r.w, "    (no drift)\n")
		return err
	case drift.StatusMissing:
		return r.writeLine("-", "(file missing from live system)", colorRed)
	case drift.StatusDrifted:
		if err := r.writeLine("-", ev.Expected, colorRed); err != nil {
			return err
		}
		return r.writeLine("+", ev.Actual, colorGreen)
	}
	return nil
}

func (r *Renderer) writeLine(prefix, content, colour string) error {
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	for _, l := range lines {
		var line string
		if r.colours {
			line = fmt.Sprintf("%s%s %s%s\n", colour, prefix, l, colorReset)
		} else {
			line = fmt.Sprintf("%s %s\n", prefix, l)
		}
		if _, err := fmt.Fprint(r.w, line); err != nil {
			return err
		}
	}
	return nil
}
