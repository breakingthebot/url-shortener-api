# URL Shortener API

Go API that stores short links in PostgreSQL, redirects visitors, tracks click counts, and now verifies changes automatically in GitHub Actions.

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

## CI
- GitHub Actions runs `go mod tidy` and `go test ./...` on every push to `main` and on every pull request.

## Deployed
Not deployed in this iteration.

## Architecture Notes
This build is the foundation: you send the API a real URL, it generates a short code, stores that mapping in PostgreSQL, and then uses the same stored record to handle redirects and count clicks. I kept the code split into small packages so the core business rules, HTTP concerns, and database work can each change independently without turning `main` into a dumping ground.

The second iteration adds CI at the point where it starts to pay off: there is already a real test suite, so every push and pull request now gets automatic verification. That gives the project a basic team workflow without changing the runtime architecture or introducing deployment complexity too early.

## Notes
- The database schema is created automatically on startup for local convenience.
- Short codes are random and collision-aware, but custom aliases are not part of this iteration yet.
- The CI workflow is intentionally small and only enforces module consistency plus the Go test suite.
