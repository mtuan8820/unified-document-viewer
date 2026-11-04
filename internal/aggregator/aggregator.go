package aggregator

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Aggregator struct {
	Cache          Cache         // optional, may be nil
	Logger         *slog.Logger  // required
	Snapshot       SnapshotStore // optional, may be nil
	SnapshotMaxAge time.Duration // required when Snapshot is set
	Sources        []Source

	// Upstream call settings. Timeout is required: 0 makes every call time out immediately.
	Timeout    time.Duration // total budget per source, including retries
	Retries    int           // extra attempts after the first one; 0 = no retry
	RetryDelay time.Duration // wait between attempts
}

type SourceResult struct {
	docs   []Document
	status SourceStatus
}

func (a *Aggregator) fetchDataFromSingleSource(ctx context.Context, vin string, index int) SourceResult {

	var docs []Document
	var fetchedAt time.Time
	var found bool
	var err error

	sourceName := a.Sources[index].Name()

	if a.Cache != nil {
		docs, fetchedAt, found, err = a.Cache.Get(ctx, vin, sourceName)

	}

	if err == nil && found {
		status := SourceStatus{sourceName, StatusOK, &fetchedAt}
		return SourceResult{docs, status}
	}

	if err != nil {
		a.Logger.Warn("Something wrong with cache", "error", err)
	}

	// Cache miss
	// We need to fetch upstream API for data
	fetchCtx, cancel := context.WithTimeout(ctx, a.Timeout)

	fetch := Retry(func(ctx context.Context) ([]Document, error) {
		return a.Sources[index].FetchDocuments(ctx, vin)
	}, a.Retries, a.RetryDelay)

	docs, err = fetch(fetchCtx)
	cancel()
	if err == nil {
		//write-through (cache.Set + snapshot.Upsert with ctx)
		// The writes run in the background, so they must not use ctx directly:
		// it is cancelled as soon as the request finishes. WithoutCancel keeps
		// the request's values but drops its cancellation; each write gets its
		// own timeout and cancels it only when that write is done.
		now := time.Now()
		bgCtx := context.WithoutCancel(ctx)

		go func() {
			if a.Cache == nil {
				return
			}
			writeCtx, cancel := context.WithTimeout(bgCtx, 2*time.Second)
			defer cancel()

			if err := a.Cache.Set(writeCtx, vin, sourceName, docs, now); err != nil {
				a.Logger.Warn("Failed to update cache", "vin", vin, "source", sourceName, "error", err)
			}
		}()

		go func() {
			if a.Snapshot == nil {
				return
			}
			writeCtx, cancel := context.WithTimeout(bgCtx, 2*time.Second)
			defer cancel()

			if err := a.Snapshot.Upsert(writeCtx, vin, sourceName, docs, now); err != nil {
				a.Logger.Warn("Failed to upsert to database", "vin", vin, "source", sourceName, "error", err)
			}
		}()

		return SourceResult{docs, SourceStatus{Source: sourceName, Status: StatusOK, FetchedAt: &now}}
	}

	// upstream failed → try snapshot
	if a.Snapshot != nil {
		snapDocs, snapAt, found, snapErr := a.Snapshot.Get(ctx, vin, sourceName, a.SnapshotMaxAge)
		if snapErr != nil {
			a.Logger.Warn("snapshot get failed", "source", sourceName, "vin", vin, "error", snapErr)
		}
		if found {
			a.Logger.Warn("upstream failed, serving stale snapshot", "source", sourceName, "vin", vin, "error", err)
			return SourceResult{snapDocs, SourceStatus{Source: sourceName, Status: StatusStale, FetchedAt: &snapAt}}
		}
	}

	// nothing usable → classify the *upstream* error
	status := StatusError
	if errors.Is(err, context.DeadlineExceeded) {
		status = StatusTimeout
	}
	a.Logger.Error("source unavailable", "source", sourceName, "vin", vin, "status", status, "error", err)
	return SourceResult{nil, SourceStatus{Source: sourceName, Status: status}}
}

func (a *Aggregator) GetDocuments(ctx context.Context, vin string, query Query) (Result, error) {
	var group errgroup.Group

	results := make([]SourceResult, len(a.Sources))

	for i := range a.Sources {

		group.Go(func() error {
			results[i] = a.fetchDataFromSingleSource(ctx, vin, i)

			return nil
		})
	}

	group.Wait()

	ret := make([]Document, 0)
	sourcesStatus := make([]SourceStatus, 0, len(a.Sources))
	anyUsable := false
	for _, r := range results {
		ret = append(ret, r.docs...)
		sourcesStatus = append(sourcesStatus, r.status)
		if r.status.Status == StatusOK || r.status.Status == StatusStale {
			anyUsable = true
		}
	}

	// Compare by the requested field first; on a tie, fall back to the
	// unique key (source, document_id) so the order is deterministic and
	// pagination never repeats or skips a document.
	sortFunc := func(x, y Document) int {
		var primary int
		if query.Sort == SortByUpdatedAt {
			primary = x.UpdatedAt.Compare(y.UpdatedAt)
		} else {
			primary = strings.Compare(x.Title, y.Title)
		}
		return cmp.Or(
			primary,
			strings.Compare(string(x.Source), string(y.Source)),
			strings.Compare(x.ID, y.ID),
		)
	}

	// Sort
	if query.Order == OrderDesc {
		slices.SortFunc(ret, reverse(sortFunc))

	} else {
		slices.SortFunc(ret, sortFunc)
	}

	// Pagination
	total := len(ret)
	size := query.Size
	if size <= 0 {
		size = 20
	}
	page := max(query.Page, 1)

	start := min((page-1)*size, total) // out-of-range page → empty slice
	end := min(start+size, total)

	result := Result{
		Documents: ret[start:end],
		Total:     total,
		Page:      page,
		HasNext:   end < total,
		Statuses:  sourcesStatus,
	}
	if len(a.Sources) > 0 && !anyUsable {
		return result, ErrAllSourcesFailed
	}
	return result, nil
}

func reverse[T any](f func(a, b T) int) func(a, b T) int {
	return func(a, b T) int {
		return -f(a, b) // Inverts -1 to 1, and 1 to -1
	}
}
