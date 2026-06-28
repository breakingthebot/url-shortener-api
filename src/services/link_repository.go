// src/services/link_repository.go
// Declares the persistence contract for storing, resolving, and inspecting shortened links.
// Connects to: src/services/link_service.go, src/services/link_repository_postgres.go
// Created: 2026-06-28

package services

import (
	"context"

	"github.com/breakingthebot/url-shortener-api/src/models"
)

// LinkRepository abstracts link persistence behind a small business-focused interface.
type LinkRepository interface {
	CreateLink(ctx context.Context, code string, originalURL string) (models.Link, error)
	GetLinkByCode(ctx context.Context, code string) (models.Link, error)
	IncrementClickCount(ctx context.Context, code string) error
	EnsureSchema(ctx context.Context) error
}
