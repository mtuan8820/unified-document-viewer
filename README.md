# unified-document-viewer

## Description
Unified Document Viewer - fetch data from 2 external system API, merge and return related documents. 

- Swagger UI is the client.
- Unified Search: Provide a single search interface where a user can enter a Vehicle Identification Number (VIN).

- Data Aggregation: The backend must make parallel requests to two mocked external APIs (a "Sales System API" and a "Service System API").


## Scope

**In scope**  
- Backend API: `GET /vehicles/{vin}/documents` returns a merged, sorted, paginated list of
  document metadata from the Sales and Service systems, with the status of each source.
- Two mock servers simulating the Sales and Service System APIs (can simulate delay and errors).
- Swagger UI as the client for trying the API.

**Out of scope**
- A frontend UI.
- Authentication and authorization.
- Downloading or viewing document contents (only metadata is returned).
- Integration with real dealership systems.

## Assumptions

- **Each VIN has a small number of documents (fewer than 500)**, so results from both
  sources are merged in memory and paginated in the backend.
- **The two external systems use different document schemas.** Each has its own adapter that
  converts its response into one common document format.
- **Document IDs are only unique within their own system**, so a document is identified by
  the pair `(source_id, source)`.
- **Documents for a vehicle change rarely.** Sales documents are created once at purchase;
  service documents every few months. A cached or saved copy is therefore usually still correct.
- **A source that doesn't respond within 2 seconds is treated as unavailable**, so one slow
  system doesn't block the whole response.

## Design decisions

- **Partial results over failure:** if one source is slow or down, the API still returns the
  other source's documents and reports each source's status (`ok`, `stale`, `timeout`, `error`).
  It returns 502 only when no data is available from either source.
- **Snapshot fallback:** every successful fetch is saved to Postgres. If a source later fails,
  the saved copy is returned and marked `stale`, but only if it is less than 30 days old
  (configurable via `SNAPSHOT_MAX_AGE`).
- **Two timestamps, two questions:** each document has `updated_at` (when it last changed in
  its source system); each source status has `fetched_at` (how fresh that source's data is in
  this response: now, cache time, or snapshot time).
- **Sorting:** by document title by default; by `updated_at` optionally.

See [docs/architecture.md](docs/architecture.md) for diagrams and details.

## Future extensions

- Document details/contents fetched on demand when the user clicks download.

## Tech Stack
- Language: Go
- Framework: Gin
- Cache: Redis
- Migrate: golang-migrate
- Observability: OpenTelemetry + jaeger
- Api spec: OpenAPI + oapi-codegen + Swagger UI
- Test: stdlib `testing`
- Database: Postgres
- Deployment: docker compose
- Logging: slog

## Getting started
### Prerequisites:
- Docker
- Go
- make

- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for `make migrate`)

### Running the project

Each step below runs in its own terminal unless noted.

1. **Start Redis, Postgres and Jaeger** (runs in the background):
   ```bash
   make up
   ```
2. **Create the database table:**
   ```bash
   make migrate
   ```
3. **Start the two mock systems:**
   ```bash
   make mock-sales     # Sales System API on :8081
   make mock-service   # Service System API on :8082
   ```
4. **Start the API server:**
   ```bash
   make run            # listens on :8080
   ```
5. **Try it out:** open Swagger UI at <http://localhost:8080/docs>, or call the API directly:
   ```bash
   curl "http://localhost:8080/vehicles/1HGCM82633A004352/documents?sort=updated_at&order=desc&page=1&size=10"
   ```

To stop the containers: `make down`. To run the tests: `make test`.

### Sample VINs

| VIN                 | Sales       | Service     | What it shows                                   |
|---------------------|-------------|-------------|-------------------------------------------------|
| `1HGCM82633A004352` | 3 documents | 3 documents | Merged list; ID `1001` exists in both systems   |
| `5YJSA1E26HF000337` | not found   | 2 documents | Documents from one source only                  |
| `3VWFE21C04M000001` | 0 documents | not found   | Known vehicle with no documents                 |

### Configuration

The server reads these environment variables. The defaults match a local setup, so nothing
needs to be set to run it. `.env.example` lists them all, but it is not loaded automatically;
export the values you want to change, e.g. `UPSTREAM_TIMEOUT=500ms make run`.

| Variable               | Default                                                         | Purpose                                         |
|------------------------|-----------------------------------------------------------------|-------------------------------------------------|
| `HTTP_ADDR`            | `:8080`                                                         | Address the API listens on                      |
| `DATABASE_URL`         | `postgres://app:app@localhost:5432/documents?sslmode=disable`   | Postgres for snapshots (also used by `make migrate`) |
| `REDIS_ADDR`           | `localhost:6379`                                                | Redis cache                                     |
| `SALES_API_URL`        | `http://localhost:8081`                                         | Sales System API                                |
| `SERVICE_API_URL`      | `http://localhost:8082`                                         | Service System API                              |
| `UPSTREAM_TIMEOUT`     | `2s`                                                            | Total time per source, including retries        |
| `UPSTREAM_RETRIES`     | `1`                                                             | Extra attempts after the first call (0 = none)  |
| `UPSTREAM_RETRY_DELAY` | `200ms`                                                         | Wait between attempts                           |
| `SNAPSHOT_MAX_AGE`     | `720h`                                                          | Oldest snapshot that may be served as `stale`   |
| `LOG_LEVEL`            | `info`                                                          | `debug`, `info`, `warn` or `error`              |

If Redis or Postgres is not running, the server still starts and logs a warning: requests skip
the cache, and the snapshot fallback is unavailable.

