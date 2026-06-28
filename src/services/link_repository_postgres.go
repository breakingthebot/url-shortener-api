// src/services/link_repository_postgres.go
// Implements PostgreSQL-backed persistence for shortened links using pgx connection pooling.
// Connects to: src/services/link_repository.go, src/main.go, PostgreSQL
// Created: 2026-06-28

package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/breakingthebot/url-shortener-api/src/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ensureLinksTableSQL = `
CREATE TABLE IF NOT EXISTS links (
	code TEXT PRIMARY KEY,
	original_url TEXT NOT NULL,
	click_count BIGINT NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// PostgresLinkRepository stores links in PostgreSQL.
type PostgresLinkRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresLinkRepository constructs a PostgreSQL-backed link repository.
func NewPostgresLinkRepository(pool *pgxpool.Pool) PostgresLinkRepository {
	return PostgresLinkRepository{pool: pool}
}

// EnsureSchema creates the required database table if it does not already exist.
func (r PostgresLinkRepository) EnsureSchema(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, ensureLinksTableSQL)
	if err != nil {
		return fmt.Errorf("ensure links table: %w", err)
	}

	return nil
}

// CreateLink inserts a new short link row and returns the stored record.
func (r PostgresLinkRepository) CreateLink(ctx context.Context, code string, originalURL string) (models.Link, error) {
	const query = `
INSERT INTO links (code, original_url)
VALUES ($1, $2)
RETURNING code, original_url, click_count, created_at;
`

	var link models.Link
	err := r.pool.QueryRow(ctx, query, code, originalURL).Scan(
		&link.Code,
		&link.OriginalURL,
		&link.ClickCount,
		&link.CreatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return models.Link{}, ErrCodeCollision
		}

		return models.Link{}, fmt.Errorf("insert link: %w", err)
	}

	return link, nil
}

// GetLinkByCode fetches a stored short link by its shortcode.
func (r PostgresLinkRepository) GetLinkByCode(ctx context.Context, code string) (models.Link, error) {
	const query = `
SELECT code, original_url, click_count, created_at
FROM links
WHERE code = $1;
`

	var link models.Link
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&link.Code,
		&link.OriginalURL,
		&link.ClickCount,
		&link.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Link{}, ErrLinkNotFound
	}

	if err != nil {
		return models.Link{}, fmt.Errorf("select link: %w", err)
	}

	return link, nil
}

// IncrementClickCount increases the stored click count for a shortcode.
func (r PostgresLinkRepository) IncrementClickCount(ctx context.Context, code string) error {
	const query = `
UPDATE links
SET click_count = click_count + 1
WHERE code = $1;
`

	commandTag, err := r.pool.Exec(ctx, query, code)
	if err != nil {
		return fmt.Errorf("update click count: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrLinkNotFound
	}

	return nil
}

// isUniqueViolation maps PostgreSQL unique constraint failures to a stable domain error.
func isUniqueViolation(err error) bool {
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) {
		return false
	}

	return pgError.Code == "23505" || strings.Contains(strings.ToLower(pgError.Message), "duplicate")
}
