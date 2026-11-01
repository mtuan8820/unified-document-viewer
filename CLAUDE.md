# CLAUDE.md

When working with this codebase, prioritize readability over cleverness. Ask clarifying questions before making architectural changes.

## Project overview

Unified Document Viewer to provide a single view of all documents related to a vehicle by connecting to two different (mocked)
dealership systems

## Architecture
- @docs/architecture.md

## Key Directories

```
.
├── cmd/server/          # Main application 
├── api/                 # API contract
├── internal/
│   ├── config/          
│   ├── httpapi/         # Gin handlers: validate VIN/query params, call aggregator, map to API types, pick 200/400/502
│   ├── aggregator/      # Core logic
│   ├── adapters/        # One package per external system, call API + map to aggregator.Document
│   ├── cache/           # Redis cache
│   └── snapshot/        # Postgres repo
├── mocks/
│   ├── sales/           # Mock sales system API
│   └── service/         # Mock service system API
├── migrations/          # Golang-migrate SQL files
└── docs/                
```

## Commands

```bash
make run                     # go run ./cmd/server
make up / make down          # docker compose: redis, postgres, jaeger (app + mocks not added yet)
make test                    # go test -race ./...
```

## What the service does

`GET /vehicles/{vin}/documents?sort=name|date&order=asc|desc&page=&size=` returns one merged, paginated list of documents for a VIN from two mocked external systems — **sales** and **service** — plus a per-source status. Contract is being written in `api/openapi.yaml` (OpenAPI 3.0.3, for oapi-codegen; avoid 3.1-only syntax like `type: [string, "null"]`).

Response shape: `{data: [Document], total_doc, page, has_next, status: [{source, status, as_of?}]}`.
- `Document`: `source_id` (string), `source` (`sales|service`), `name`, `date`, `type`. A document is identified by the pair `(source_id, source)` — ids can collide across systems.
- Per-source `status`: `ok` (fresh), `stale` (upstream failed, served from snapshot, includes `as_of`), `timeout`, `error` (5xx/DNS/conn refused).
- 200 if any source yields data (or both succeed with zero docs); 400 invalid VIN; 502 only when both sources fail and no usable snapshot exists.

## Architecture rules.
Responsibility split (keep it this way):
- **Adapters** only call their external API and normalize to `aggregator.Document`. They implement `aggregator.Source`, whose methods must be exported (an unexported interface method can't be implemented outside the package). The design calls for `Name() string` alongside `FetchDocuments`.
- **Aggregator** owns everything else: parallel fan-out, per-source timeout, cache, snapshot fallback, merge, sort, and per-source status. Use plain `errgroup.Group` (not `WithContext`) so one failing source doesn't cancel the other; write each source's result to its own slot rather than appending to a shared slice from goroutines.
- Per source flow: Redis hit -> return; miss -> call upstream with timeout -> on success update Redis + upsert snapshot; on failure read snapshot newer than `SNAPSHOT_MAX_AGE` -> `stale`, else `timeout`/`error`.
- Merge everything, then sort and paginate in the backend (assumption: total docs per VIN is small).
- The aggregator should keep its own `Document` type; convert to/from oapi-codegen types in the handler.

## Persistence

Single Postgres table `snapshot(vin varchar(17), source <enum>, data jsonb, fetched_at timestamptz, PK(vin, source))` storing **normalized** document metadata (not raw upstream payloads, no file contents). Upsert only when the incoming `fetched_at` is newer. No cleanup job — rows are overwritten on the next successful fetch; staleness is enforced at read time with `fetched_at > now() - SNAPSHOT_MAX_AGE`. Migrations go in `migrations/` (golang-migrate).

