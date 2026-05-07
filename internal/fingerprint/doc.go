// Package fingerprint generates and compares content fingerprints for
// driftwatch. It provides a thin, algorithm-agnostic wrapper around
// standard hashing primitives so that the rest of the codebase never
// depends directly on a specific hash function.
//
// Usage:
//
//	fp := fingerprint.Sum(fileContent)
//	if !fp.Equal(stored) {
//		// content has changed
//	}
//
// Fingerprints are serialised as "algorithm:hexdigest" strings and can
// be round-tripped through [Parse].
package fingerprint
