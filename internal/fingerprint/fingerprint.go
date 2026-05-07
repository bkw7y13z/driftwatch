// Package fingerprint provides utilities for generating and comparing
// content fingerprints used to detect configuration drift.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Algorithm represents a hashing algorithm.
type Algorithm string

const (
	SHA256 Algorithm = "sha256"
)

// Result holds the fingerprint of a piece of content.
type Result struct {
	Algorithm Algorithm
	Hex       string
}

// String returns the algorithm-prefixed fingerprint string.
func (r Result) String() string {
	return string(r.Algorithm) + ":" + r.Hex
}

// Equal reports whether two Results represent the same content.
func (r Result) Equal(other Result) bool {
	return r.Algorithm == other.Algorithm && r.Hex == other.Hex
}

// Sum computes a SHA-256 fingerprint of the given content.
// Leading/trailing whitespace is normalised before hashing.
func Sum(content string) Result {
	normalised := strings.TrimSpace(content)
	h := sha256.Sum256([]byte(normalised))
	return Result{
		Algorithm: SHA256,
		Hex:       hex.EncodeToString(h[:]),
	}
}

// Parse parses a fingerprint string of the form "algorithm:hex".
// It returns ok=false if the format is unrecognised.
func Parse(s string) (Result, bool) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return Result{}, false
	}
	alg := Algorithm(parts[0])
	if alg != SHA256 {
		return Result{}, false
	}
	return Result{Algorithm: alg, Hex: parts[1]}, true
}
