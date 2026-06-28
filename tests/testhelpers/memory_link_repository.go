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
	clickEvents    map[string][]models.ClickEvent
	collisionCodes map[string]bool
}

// NewMemoryLinkRepository constructs an empty in-memory repository.
func NewMemoryLinkRepository() *MemoryLinkRepository {
	return &MemoryLinkRepository{
		links:          map[string]models.Link{},
		clickEvents:    map[string][]models.ClickEvent{},
		collisionCodes: map[string]bool{},
	}
}

// CreateLink stores a link unless the code already exists or has been marked to collide once.
func (r *MemoryLinkRepository) CreateLink(_ context.Context, code string, originalURL string, expiresAt *time.Time) (models.Link, error) {
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
		ExpiresAt:   expiresAt,
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

// GetLinkByOriginalURL returns a stored link for an original URL when it exists.
func (r *MemoryLinkRepository) GetLinkByOriginalURL(_ context.Context, originalURL string) (models.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, link := range r.links {
		if link.OriginalURL == originalURL {
			return link, nil
		}
	}

	return models.Link{}, services.ErrLinkNotFound
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

// RecordClickEvent appends a click event for the stored link.
func (r *MemoryLinkRepository) RecordClickEvent(_ context.Context, code string, clickedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[code]; !exists {
		return services.ErrLinkNotFound
	}

	r.clickEvents[code] = append(r.clickEvents[code], models.ClickEvent{
		Code:      code,
		ClickedAt: clickedAt,
	})
	return nil
}

// ListClickEvents returns recent click events for a stored link ordered newest first.
func (r *MemoryLinkRepository) ListClickEvents(_ context.Context, code string, limit int) ([]models.ClickEvent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.links[code]; !exists {
		return nil, services.ErrLinkNotFound
	}

	events := r.clickEvents[code]
	if len(events) == 0 {
		return []models.ClickEvent{}, nil
	}

	results := make([]models.ClickEvent, 0, min(limit, len(events)))
	for index := len(events) - 1; index >= 0 && len(results) < limit; index-- {
		results = append(results, events[index])
	}

	return results, nil
}

// SoftDeleteLink timestamps the stored link as deleted.
func (r *MemoryLinkRepository) SoftDeleteLink(_ context.Context, code string, deletedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, exists := r.links[code]
	if !exists {
		return services.ErrLinkNotFound
	}

	link.DeletedAt = &deletedAt
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

// min returns the lower of two integers.
func min(left int, right int) int {
	if left < right {
		return left
	}

	return right
}
