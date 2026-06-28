// src/services/link_service.go
// Coordinates URL validation, shortcode generation, repository writes, and click tracking.
// Connects to: src/utils/validation/url.go, src/utils/shortcode/generator.go, src/services/link_repository.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/breakingthebot/url-shortener-api/src/models"
	"github.com/breakingthebot/url-shortener-api/src/utils/shortcode"
	"github.com/breakingthebot/url-shortener-api/src/utils/validation"
)

const maxCreateAttempts = 5

// LinkService contains the business logic for shortened links.
type LinkService struct {
	repository LinkRepository
	generator  shortcode.Generator
	logger     *slog.Logger
}

// NewLinkService builds a link service with its storage dependency and shortcode generator.
func NewLinkService(repository LinkRepository, generator shortcode.Generator, logger *slog.Logger) LinkService {
	return LinkService{
		repository: repository,
		generator:  generator,
		logger:     logger,
	}
}

// CreateShortLink validates inputs, reuses duplicates when possible, and persists a short link when needed.
func (s LinkService) CreateShortLink(ctx context.Context, originalURL string, customCode string) (models.Link, bool, error) {
	normalizedURL, err := validation.NormalizeURL(originalURL)
	if err != nil {
		s.logger.Warn("invalid original url", "error", err)
		return models.Link{}, false, fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}

	normalizedCustomCode, hasCustomCode, err := validation.NormalizeCustomCode(customCode)
	if err != nil {
		s.logger.Warn("invalid custom code", "error", err)
		return models.Link{}, false, fmt.Errorf("%w: %s", ErrInvalidCustomCode, err.Error())
	}

	if hasCustomCode {
		existingLinkByCode, lookupErr := s.repository.GetLinkByCode(ctx, normalizedCustomCode)
		if lookupErr == nil {
			if existingLinkByCode.OriginalURL == normalizedURL {
				s.logger.Info("existing short link reused by custom code", "code", existingLinkByCode.Code)
				return existingLinkByCode, false, nil
			}

			return models.Link{}, false, fmt.Errorf("%w: %s", ErrCustomCodeUnavailable, normalizedCustomCode)
		}

		if !errors.Is(lookupErr, ErrLinkNotFound) {
			return models.Link{}, false, fmt.Errorf("get link by custom code: %w", lookupErr)
		}
	}

	existingLinkByURL, err := s.repository.GetLinkByOriginalURL(ctx, normalizedURL)
	if err == nil {
		if !hasCustomCode || existingLinkByURL.Code == normalizedCustomCode {
			s.logger.Info("existing short link reused by original url", "code", existingLinkByURL.Code)
			return existingLinkByURL, false, nil
		}

		return models.Link{}, false, fmt.Errorf("%w: existing code %s", ErrURLAlreadyShortened, existingLinkByURL.Code)
	}

	if !errors.Is(err, ErrLinkNotFound) {
		return models.Link{}, false, fmt.Errorf("get link by original url: %w", err)
	}

	if hasCustomCode {
		link, createErr := s.repository.CreateLink(ctx, normalizedCustomCode, normalizedURL)
		if createErr != nil {
			if errors.Is(createErr, ErrCodeCollision) {
				return models.Link{}, false, fmt.Errorf("%w: %s", ErrCustomCodeUnavailable, normalizedCustomCode)
			}

			return models.Link{}, false, fmt.Errorf("create short link with custom code: %w", createErr)
		}

		s.logger.Info("short link created with custom code", "code", link.Code)
		return link, true, nil
	}

	for attempt := 0; attempt < maxCreateAttempts; attempt++ {
		code, generateErr := s.generator.Generate()
		if generateErr != nil {
			return models.Link{}, false, fmt.Errorf("generate shortcode: %w", generateErr)
		}

		link, createErr := s.repository.CreateLink(ctx, code, normalizedURL)
		if errors.Is(createErr, ErrCodeCollision) {
			s.logger.Info("retrying after shortcode collision", "code", code, "attempt", attempt+1)
			continue
		}

		if createErr != nil {
			return models.Link{}, false, fmt.Errorf("create short link: %w", createErr)
		}

		s.logger.Info("short link created", "code", link.Code)
		return link, true, nil
	}

	return models.Link{}, false, fmt.Errorf("create short link: %w", ErrCodeCollision)
}

// ResolveShortLink returns the original URL and increments its click count.
func (s LinkService) ResolveShortLink(ctx context.Context, code string) (string, error) {
	link, err := s.repository.GetLinkByCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("get link by code: %w", err)
	}

	if err := s.repository.IncrementClickCount(ctx, code); err != nil {
		return "", fmt.Errorf("increment click count: %w", err)
	}

	s.logger.Info("short link resolved", "code", code)
	return link.OriginalURL, nil
}

// GetLinkStats returns the current stored state for a shortcode.
func (s LinkService) GetLinkStats(ctx context.Context, code string) (models.Link, error) {
	link, err := s.repository.GetLinkByCode(ctx, code)
	if err != nil {
		return models.Link{}, fmt.Errorf("get link stats: %w", err)
	}

	return link, nil
}
