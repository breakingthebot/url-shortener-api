// tests/testhelpers/memory_link_repository.go
// Provides an in-memory repository double for isolated service and handler tests.
// Connects to: tests/services/link_service_test.go, tests/components/httpapi/link_handler_test.go
// Created: 2026-06-28

package testhelpers

import (
	"context"
	"sync"
	"time"

	"github.com/breakingthebot/url-shortener-api/src/models"
	"github.com/breakingthebot/url-shortener-api/src/services"
)

// MemoryLinkRepository is a concurrency-safe in-memory implementation of the repository contract.
type MemoryLinkRepository struct {
	mu             sync.Mutex
	links          map[string]models.Link
	collisionCodes map[string]bool
}

// NewMemoryLinkRepository constructs an empty in-memory repository.
func NewMemoryLinkRepository() *MemoryLinkRepository {
	return &MemoryLinkRepository{
		links:          map[string]models.Link{},
		collisionCodes: map[string]bool{},
	}
}

// CreateLink stores a link unless the code already exists or has been marked to collide once.
func (r *MemoryLinkRepository) CreateLink(_ context.Context, code string, originalURL string) (models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.collisionCodes[code] {
		delete(r.collisionCodes, code)
		return models.Link{}, services.ErrCodeCollision
	}

	if _, exists := r.links[code]; exists {
		return models.Link{}, services.ErrCodeCollision
	}

	link := models.Link{
		Code:        code,
		OriginalURL: originalURL,
		ClickCount:  0,
		CreatedAt:   time.Now().UTC(),
	}
	r.links[code] = link

	return link, nil
}

// GetLinkByCode returns a stored link when it exists.
func (r *MemoryLinkRepository) GetLinkByCode(_ context.Context, code string) (models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, exists := r.links[code]
	if !exists {
		return models.Link{}, services.ErrLinkNotFound
	}

	return link, nil
}

// IncrementClickCount increases the click counter for the stored link.
func (r *MemoryLinkRepository) IncrementClickCount(_ context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, exists := r.links[code]
	if !exists {
		return services.ErrLinkNotFound
	}

	link.ClickCount++
	r.links[code] = link
	return nil
}

// EnsureSchema satisfies the repository interface for non-database tests.
func (r *MemoryLinkRepository) EnsureSchema(_ context.Context) error {
	return nil
}

// MarkNextCollision causes the next create attempt for a code to collide.
func (r *MemoryLinkRepository) MarkNextCollision(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.collisionCodes[code] = true
}
