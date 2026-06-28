# URL Shortener API

Go API that stores short links in PostgreSQL, supports optional custom aliases, reuses duplicate URLs, redirects visitors, tracks click counts, and verifies changes automatically in GitHub Actions.

## Stack
- Go 1.26
- Standard library `net/http` and `log/slog`
- `github.com/jackc/pgx/v5` for PostgreSQL
- PostgreSQL

## Setup
1. Install Go 1.26 or later.
2. Start a PostgreSQL instance.
3. Copy `.env.example` values into your shell environment or a local `.env` loader.
4. Run `go mod tidy`.

## Environment Variables
- `APP_ENV`
- `APP_HOST`
- `APP_PORT`
- `DATABASE_URL`
- `SHORT_CODE_LENGTH`

## Running Locally
```bash
go test ./...
go run ./src
```

## API
- `POST /api/v1/links` accepts `original_url` and optional `custom_code`.
- Re-submitting the same `original_url` returns the existing short link instead of creating a duplicate row.
- Requesting a `custom_code` that is already assigned to a different URL returns `409 Conflict`.

## CI
- GitHub Actions runs `go mod tidy` and `go test ./...` on every push to `main` and on every pull request.

## Deployed
Not deployed in this iteration.

## Architecture Notes
This build is the foundation: you send the API a real URL, it generates a short code, stores that mapping in PostgreSQL, and then uses the same stored record to handle redirects and count clicks. I kept the code split into small packages so the core business rules, HTTP concerns, and database work can each change independently without turning `main` into a dumping ground.

The second iteration added CI once the test suite existed, and this third one makes the API more practical for real usage. Teams usually do not want multiple rows for the same destination URL, and they often need branded or memorable aliases for links they share externally.

I handled that by keeping the same layered structure and extending the create flow instead of adding special-case logic in the handler. The service now decides when to reuse an existing record, when to honor a requested alias, and when to return a conflict because the requested code cannot safely map to the submitted URL.

## Notes
- The database schema is created automatically on startup for local convenience.
- Random short codes are still collision-aware, and custom aliases are now validated to only allow route-safe characters.
- The CI workflow is intentionally small and only enforces module consistency plus the Go test suite.
