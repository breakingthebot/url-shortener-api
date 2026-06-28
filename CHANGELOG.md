# Changelog

## [0.3.0] - 2026-06-28
- Add optional `custom_code` support to link creation requests.
- Reuse existing short links when the same original URL is submitted again.
- Return `409 Conflict` when a requested custom code is already tied to a different URL.
- Add validation and handler tests for aliasing and duplicate URL reuse.

## [0.2.0] - 2026-06-28
- Add a GitHub Actions CI workflow that runs `go mod tidy` and `go test ./...` on pushes to `main` and on pull requests.
- Update the README to document the automated verification workflow.

## [0.1.0] - 2026-06-28
- Create the initial Go project structure for a PostgreSQL-backed URL shortener API.
- Add HTTP endpoints for health checks, short link creation, redirect handling, and stats lookup.
- Add service, validation, and handler tests plus the initial MIT license and setup documentation.
