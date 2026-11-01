package aggregator

import "time"

type Status string

const (
	StatusOK      Status = "ok"      // fresh from upstream or from cache
	StatusStale   Status = "stale"   // upstream failed, served from snapshot
	StatusTimeout Status = "timeout" // upstream timed out, no usable snapshot
	StatusError   Status = "error"   // upstream failed (5xx, DNS, conn refused), no usable snapshot
)

// SourceStatus reports the outcome for one source.
type SourceStatus struct {
	Source SourceName
	Status Status
	// FetchedAt is when our backend obtained the data: now (fresh), cache
	// time, or snapshot time. Nil for StatusTimeout and StatusError.
	FetchedAt *time.Time
}
