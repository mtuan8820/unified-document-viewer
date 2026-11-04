package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	agg "github.com/mtuan8820/unified-document-viewer/internal/aggregator"
	"github.com/redis/go-redis/v9"
)

// Compile-time check that DocumentsCache satisfies the aggregator's interface.
var _ agg.Cache = (*DocumentsCache)(nil)

type DocumentsCache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *DocumentsCache {
	return &DocumentsCache{client: client, ttl: ttl}
}

// entry is what we store in Redis. FetchedAt is kept alongside the docs
// because the aggregator reports it as SourceStatus.fetched_at on a hit.
type entry struct {
	FetchedAt time.Time      `json:"fetched_at"`
	Docs      []agg.Document `json:"docs"`
}

func key(vin string, src agg.SourceName) string {
	return "docs:" + vin + ":" + string(src)
}

func (c *DocumentsCache) Get(ctx context.Context, vin string, src agg.SourceName) (docs []agg.Document, fetchedAt time.Time, found bool, err error) {
	raw, err := c.client.Get(ctx, key(vin, src)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, time.Time{}, false, nil // key missing: a miss, not an error
	}
	if err != nil {
		return nil, time.Time{}, false, err
	}

	var e entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, time.Time{}, false, err
	}
	return e.Docs, e.FetchedAt, true, nil
}

func (c *DocumentsCache) Set(ctx context.Context, vin string, src agg.SourceName, docs []agg.Document, fetchedAt time.Time) error {
	raw, err := json.Marshal(entry{FetchedAt: fetchedAt, Docs: docs})
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(vin, src), raw, c.ttl).Err()
}
