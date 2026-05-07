// Package suppress provides temporary suppression of drift alerts for
// known maintenance windows or accepted exceptions.
//
// A Store holds a set of time-bounded entries keyed by service name and
// file path. When an entry is active (i.e. its expiry is in the future)
// the corresponding drift event is silenced upstream in the notification
// and alerting pipeline.
//
// Entries are never automatically removed; callers should invoke Purge
// periodically — for example from the scheduler — to reclaim memory.
package suppress
