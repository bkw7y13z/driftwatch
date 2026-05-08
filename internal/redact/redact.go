// Package redact provides utilities for scrubbing sensitive values from
// configuration content before it is logged, audited, or compared.
package redact

import (
	"regexp"
	"strings"
)

// DefaultPatterns is the set of key patterns whose values are redacted by
// default (case-insensitive substring match on the key name).
var DefaultPatterns = []string{
	"password",
	"secret",
	"token",
	"api_key",
	"apikey",
	"private_key",
	"credentials",
}

const redactedPlaceholder = "[REDACTED]"

// Redactor scrubs sensitive key/value pairs from YAML-like text.
type Redactor struct {
	patterns []*regexp.Regexp
}

// New returns a Redactor that masks values whose keys match any of the
// supplied patterns (case-insensitive). If patterns is empty, DefaultPatterns
// is used.
func New(patterns []string) (*Redactor, error) {
	if len(patterns) == 0 {
		patterns = DefaultPatterns
	}
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(p))
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return &Redactor{patterns: compiled}, nil
}

// Apply returns a copy of content with sensitive values replaced by
// [REDACTED]. It operates line-by-line and handles "key: value" style lines.
func (r *Redactor) Apply(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = r.redactLine(line)
	}
	return strings.Join(lines, "\n")
}

// IsSensitiveKey reports whether the given key name matches any pattern.
func (r *Redactor) IsSensitiveKey(key string) bool {
	for _, re := range r.patterns {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

// redactLine replaces the value portion of a "key: value" line if the key is
// sensitive. Lines that do not match the pattern are returned unchanged.
func (r *Redactor) redactLine(line string) string {
	// Match optional leading whitespace, a key, a colon, and a value.
	colon := strings.Index(line, ":")
	if colon < 0 {
		return line
	}
	key := strings.TrimSpace(line[:colon])
	if !r.IsSensitiveKey(key) {
		return line
	}
	leading := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return leading + key + ": " + redactedPlaceholder
}
