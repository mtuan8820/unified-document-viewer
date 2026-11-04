package aggregator

import "errors"

// ErrAllSourcesFailed is returned by GetDocuments when no source produced
// usable data (every status is timeout or error). The Result is still
// returned alongside it, so the caller can report each source's status.
var ErrAllSourcesFailed = errors.New("aggregator: all sources failed")
