// Package watcher provides filesystem-level file reading for drift detection.
//
// A Watcher is constructed with one or more base directory paths. When asked
// to read a relative file path it searches each base directory in order,
// returning the first match. Files that cannot be found in any base directory
// are represented as missing FileState values rather than hard errors, so the
// drift pipeline can distinguish between "file differs" and "file absent".
//
// Typical usage:
//
//	w := watcher.New([]string{"/etc/myservice", "/opt/myservice"}, logger)
//	states, err := w.ReadAll(ctx, declaredPaths)
//	live := watcher.ToLiveContent(states)
//	// pass live to drift.Detector.CompareAll
package watcher
