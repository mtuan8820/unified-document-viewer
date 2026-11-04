package aggregator

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"testing"
	"time"
)

// For each source there are 5 outputs:
// 1. Cache hit -> return cached documents, status OK
// 2. Cache miss -> fetch upstream success -> return documents, status OK (and write cache + snapshot)
// 3. Cache miss -> fetch upstream fail (error/timeout) -> snapshot found -> return snapshot, status STALE
// 4. Cache miss -> fetch upstream timeout -> no usable snapshot -> status TIMEOUT
// 5. Cache miss -> fetch upstream error -> no usable snapshot -> status ERROR
//
// "No usable snapshot" means no row OR a row older than SnapshotMaxAge. The
// snapshot store applies the age filter, so both look like found=false here;
// the age check itself is tested in the snapshot package.

// We need to test GetDocuments.
// It runs one goroutine per source, and the sources don't affect each other,
// so we can mix outputs in one case to reduce the number of cases
// instead of trying all the combinations.

// For short, service system is A, sales system is B.

// We need to test these cases:

// A 1, B 2
// A 2, B 3
// A 3, B 4
// A 4, B 5 -> both failed -> ErrAllSourcesFailed
// A 5, B 1
// A 2, B 2 but zero docs on both -> no error, return empty list of docs

const testVIN = "1HGCM82633A004352"

var (
	cacheAt     = time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	snapshotAt  = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	errUpstream = errors.New("upstream returned 503")
	// Used by cache-hit setups: if the aggregator calls upstream anyway,
	// the source fails and the test catches it through the status.
	errNotCalled = errors.New("upstream should not be called")
)

// ---- per-source setup ----

type entry struct {
	docs []Document
	at   time.Time
}

// setup describes how one source behaves in a test case.
type setup struct {
	cache       *entry // nil = cache miss
	upstream    []Document
	upstreamErr error
	slow        bool   // upstream blocks until the timeout
	snapshot    *entry // nil = no usable snapshot
}

func cacheHit(docs ...Document) setup {
	return setup{cache: &entry{docs, cacheAt}, upstreamErr: errNotCalled}
}

func fetchOK(docs ...Document) setup {
	return setup{upstream: docs}
}

func staleSnapshot(docs ...Document) setup {
	return setup{upstreamErr: errUpstream, snapshot: &entry{docs, snapshotAt}}
}

func timesOut() setup {
	return setup{slow: true}
}

func fails() setup {
	return setup{upstreamErr: errUpstream}
}

func doc(src SourceName, id string) Document {
	return Document{ID: id, Source: src, Title: "doc " + id, Type: ""}
}

// ---- expected output per source ----

// wantAt says which time FetchedAt should hold.
type wantAt int

const (
	atNone     wantAt = iota // nil (timeout / error)
	atCache                  // cacheAt
	atNow                    // time of the call (fresh fetch)
	atSnapshot               // snapshotAt
)

type want struct {
	status Status
	at     wantAt
}

type mockSource struct {
	name SourceName
	s    setup
}

func (f *mockSource) Name() SourceName { return f.name }

func (f *mockSource) FetchDocuments(ctx context.Context, vin string) ([]Document, error) {
	if f.s.slow {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.s.upstream, f.s.upstreamErr
}

// mockCache entries are only read during the test, so concurrent Gets are safe.
// Set runs in a background goroutine, so it reports through a channel.
type mockCache struct {
	entries map[SourceName]entry
	writes  chan SourceName
}

func (c *mockCache) Get(ctx context.Context, vin string, src SourceName) ([]Document, time.Time, bool, error) {
	e, ok := c.entries[src]
	return e.docs, e.at, ok, nil
}

func (c *mockCache) Set(ctx context.Context, vin string, src SourceName, docs []Document, fetchedAt time.Time) error {
	c.writes <- src
	return nil
}

type fakeSnapshot struct {
	entries map[SourceName]entry
	writes  chan SourceName
}

func (s *fakeSnapshot) Get(ctx context.Context, vin string, src SourceName, maxAge time.Duration) ([]Document, time.Time, bool, error) {
	e, ok := s.entries[src]
	return e.docs, e.at, ok, nil
}

func (s *fakeSnapshot) Upsert(ctx context.Context, vin string, src SourceName, docs []Document, fetchedAt time.Time) error {
	s.writes <- src
	return nil
}

// ---- test ----

func TestGetDocuments(t *testing.T) {
	tests := []struct {
		name        string
		service     setup // A
		sales       setup // B
		wantService want
		wantSales   want
		wantDocs    []string // "source/id"
		wantErr     error
	}{
		{
			name:        "A1 B2: service cache hit, sales fresh fetch",
			service:     cacheHit(doc(SourceService, "V1")),
			sales:       fetchOK(doc(SourceSales, "S1")),
			wantService: want{StatusOK, atCache},
			wantSales:   want{StatusOK, atNow},
			wantDocs:    []string{"sales/S1", "service/V1"},
		},
		{
			name:        "A2 B3: service fresh fetch, sales stale",
			service:     fetchOK(doc(SourceService, "V1")),
			sales:       staleSnapshot(doc(SourceSales, "S-old")),
			wantService: want{StatusOK, atNow},
			wantSales:   want{StatusStale, atSnapshot},
			wantDocs:    []string{"sales/S-old", "service/V1"},
		},
		{
			name:        "A3 B4: service stale, sales timeout",
			service:     staleSnapshot(doc(SourceService, "V-old")),
			sales:       timesOut(),
			wantService: want{StatusStale, atSnapshot},
			wantSales:   want{StatusTimeout, atNone},
			wantDocs:    []string{"service/V-old"},
		},
		{
			name:        "A4 B5: service timeout, sales error: all failed",
			service:     timesOut(),
			sales:       fails(),
			wantService: want{StatusTimeout, atNone},
			wantSales:   want{StatusError, atNone},
			wantDocs:    nil,
			wantErr:     ErrAllSourcesFailed,
		},
		{
			name:        "A5 B1: service error, sales cache hit",
			service:     fails(),
			sales:       cacheHit(doc(SourceSales, "S1")),
			wantService: want{StatusError, atNone},
			wantSales:   want{StatusOK, atCache},
			wantDocs:    []string{"sales/S1"},
		},
		{
			name:        "A2 B2: both fresh fetch, zero docs",
			service:     fetchOK(),
			sales:       fetchOK(),
			wantService: want{StatusOK, atNow},
			wantSales:   want{StatusOK, atNow},
			wantDocs:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := &mockCache{entries: map[SourceName]entry{}, writes: make(chan SourceName, 10)}
			snap := &fakeSnapshot{entries: map[SourceName]entry{}, writes: make(chan SourceName, 10)}
			for src, s := range map[SourceName]setup{SourceService: tt.service, SourceSales: tt.sales} {
				if s.cache != nil {
					cache.entries[src] = *s.cache
				}
				if s.snapshot != nil {
					snap.entries[src] = *s.snapshot
				}
			}

			agg := &Aggregator{
				Cache:          cache,
				Logger:         slog.New(slog.DiscardHandler),
				Snapshot:       snap,
				SnapshotMaxAge: time.Hour,
				Sources: []Source{
					&mockSource{SourceService, tt.service},
					&mockSource{SourceSales, tt.sales},
				},
				Timeout: 50 * time.Millisecond,
			}

			before := time.Now()
			got, err := agg.GetDocuments(context.Background(), testVIN, Query{Sort: SortByTitle, Order: OrderAsc})
			after := time.Now()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}

			checkStatus(t, got.Statuses, SourceService, tt.wantService, before, after)
			checkStatus(t, got.Statuses, SourceSales, tt.wantSales, before, after)

			if gotDocs := docKeys(got.Documents); !slices.Equal(gotDocs, tt.wantDocs) {
				t.Errorf("docs = %v, want %v", gotDocs, tt.wantDocs)
			}

			// Only a fresh fetch writes to cache and snapshot.
			var wantWrites []SourceName
			if tt.wantService.at == atNow {
				wantWrites = append(wantWrites, SourceService)
			}
			if tt.wantSales.at == atNow {
				wantWrites = append(wantWrites, SourceSales)
			}
			checkWrites(t, "cache", cache.writes, wantWrites)
			checkWrites(t, "snapshot", snap.writes, wantWrites)
		})
	}
}

// ---- helpers ----

func checkStatus(t *testing.T, statuses []SourceStatus, src SourceName, w want, before, after time.Time) {
	t.Helper()

	i := slices.IndexFunc(statuses, func(s SourceStatus) bool { return s.Source == src })
	if i < 0 {
		t.Errorf("%s: no status returned", src)
		return
	}
	got := statuses[i]

	if got.Status != w.status {
		t.Errorf("%s: status = %q, want %q", src, got.Status, w.status)
	}

	if w.at == atNone {
		if got.FetchedAt != nil {
			t.Errorf("%s: fetched_at = %v, want nil", src, *got.FetchedAt)
		}
		return
	}
	if got.FetchedAt == nil {
		t.Errorf("%s: fetched_at is nil, want a time", src)
		return
	}

	at := *got.FetchedAt
	switch w.at {
	case atCache:
		if !at.Equal(cacheAt) {
			t.Errorf("%s: fetched_at = %v, want cache time %v", src, at, cacheAt)
		}
	case atSnapshot:
		if !at.Equal(snapshotAt) {
			t.Errorf("%s: fetched_at = %v, want snapshot time %v", src, at, snapshotAt)
		}
	case atNow:
		if at.Before(before) || at.After(after) {
			t.Errorf("%s: fetched_at = %v, want between %v and %v", src, at, before, after)
		}
	}
}

// docKeys returns sorted "source/id" keys, because an id alone can collide across sources.
func docKeys(docs []Document) []string {
	var keys []string
	for _, d := range docs {
		keys = append(keys, string(d.Source)+"/"+d.ID)
	}
	slices.Sort(keys)
	return keys
}

// checkWrites waits for the background writes, since they finish after GetDocuments returns.
func checkWrites(t *testing.T, store string, writes <-chan SourceName, want []SourceName) {
	t.Helper()

	var got []SourceName
	for range want {
		select {
		case src := <-writes:
			got = append(got, src)
		case <-time.After(time.Second):
			t.Errorf("%s: timed out waiting for writes, got %v, want %v", store, got, want)
			return
		}
	}

	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("%s writes = %v, want %v", store, got, want)
	}
}
