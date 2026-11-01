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
  the saved copy is returned and marked `stale` with its timestamp, but only if it is less
  than 30 days old (configurable via `SNAPSHOT_MAX_AGE`).
- **Sorting:** by document name by default; by date optionally.

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

### Running the project


