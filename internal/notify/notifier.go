package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Level represents the severity of a drift notification.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Event holds the details of a single drift notification.
type Event struct {
	Timestamp time.Time
	Level     Level
	Service   string
	Message   string
}

// Notifier sends drift events to one or more outputs.
type Notifier struct {
	out io.Writer
}

// New creates a Notifier that writes to the given writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stdout
	}
	return &Notifier{out: w}
}

// Notify emits an Event for each drifted or missing file found in the report.
// It returns the number of events emitted.
func (n *Notifier) Notify(service string, report *drift.Report) (int, error) {
	if report == nil {
		return 0, fmt.Errorf("notify: nil report for service %q", service)
	}

	count := 0
	for _, result := range report.Results {
		var lvl Level
		var msg string

		switch result.Status {
		case drift.StatusMatch:
			continue
		case drift.StatusMissing:
			lvl = LevelError
			msg = fmt.Sprintf("file %q is missing from the running service", result.Path)
		case drift.StatusDrifted:
			lvl = LevelWarn
			msg = fmt.Sprintf("file %q has drifted from declared state", result.Path)
		default:
			lvl = LevelInfo
			msg = fmt.Sprintf("file %q has unknown status %q", result.Path, result.Status)
		}

		ev := Event{
			Timestamp: time.Now().UTC(),
			Level:     lvl,
			Service:   service,
			Message:   msg,
		}
		if err := n.write(ev); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (n *Notifier) write(ev Event) error {
	_, err := fmt.Fprintf(n.out, "%s [%s] service=%s %s\n",
		ev.Timestamp.Format(time.RFC3339), ev.Level, ev.Service, ev.Message)
	return err
}
