package aggregator

import (
	"context"
	"time"
)

type SnapshotStore interface {
	// returns found=false if no row, or if the row is older than maxAge
	Get(ctx context.Context, vin string, src SourceName, maxAge time.Duration) (docs []Document, fetchedAt time.Time, found bool, err error)
	Upsert(ctx context.Context, vin string, src SourceName, docs []Document, fetchedAt time.Time) error
}
