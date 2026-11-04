package snapshot

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	agg "github.com/mtuan8820/unified-document-viewer/internal/aggregator"
)

// make sure Store implement interface SnapshotStore
var _ agg.SnapshotStore = (*Store)(nil)

// Store keeps the last successful document list per (vin, source) in the
// snapshot table. Data is the normalized []Document, stored as jsonb.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Get returns found=false when there is no row, or when the row is older
// than maxAge (staleness is enforced here, at read time).
func (s *Store) Get(ctx context.Context, vin string, src agg.SourceName, maxAge time.Duration) (docs []agg.Document, fetchedAt time.Time, found bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT data, fetched_at
		FROM snapshot
		WHERE vin = $1
		  AND source = $2::source_name
		  AND fetched_at > now() - make_interval(secs => $3)`,
		vin, string(src), maxAge.Seconds(),
	).Scan(&docs, &fetchedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, time.Time{}, false, nil
	}
	if err != nil {
		return nil, time.Time{}, false, err
	}
	return docs, fetchedAt, true, nil
}

// Upsert writes the snapshot, but only if fetchedAt is newer than what is
// stored, so a slow, older write can never overwrite fresher data.
func (s *Store) Upsert(ctx context.Context, vin string, src agg.SourceName, docs []agg.Document, fetchedAt time.Time) error {
	// A nil slice would be encoded as SQL NULL; zero docs is a valid
	// snapshot, so store it as an empty JSON array instead.
	if docs == nil {
		docs = []agg.Document{}
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO snapshot (vin, source, data, fetched_at)
		VALUES ($1, $2::source_name, $3, $4)
		ON CONFLICT (vin, source) DO UPDATE
		SET data = EXCLUDED.data,
		    fetched_at = EXCLUDED.fetched_at
		WHERE snapshot.fetched_at < EXCLUDED.fetched_at`,
		vin, string(src), docs, fetchedAt,
	)
	return err
}
