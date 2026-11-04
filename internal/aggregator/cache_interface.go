package aggregator

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, vin string, src SourceName) (docs []Document, fetchedAt time.Time, found bool, err error)
	Set(ctx context.Context, vin string, src SourceName, docs []Document, fetchedAt time.Time) error
}
