// tests/services/link_service_test.go
// Verifies link service behavior around URL validation, collision retries, and click counting.
// Connects to: src/services/link_service.go, tests/testhelpers/memory_link_repository.go
// Created: 2026-06-28

package services_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/breakingthebot/url-shortener-api/src/services"
	"github.com/breakingthebot/url-shortener-api/src/utils/shortcode"
	"github.com/breakingthebot/url-shortener-api/tests/testhelpers"
)

// TestCreateShortLinkRejectsInvalidURL confirms invalid URLs are blocked before persistence.
func TestCreateShortLinkRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	service := services.NewLinkService(
		testhelpers.NewMemoryLinkRepository(),
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	_, _, err := service.CreateShortLink(context.Background(), "not-a-url", "", "")
	if !errors.Is(err, services.ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

// TestCreateShortLinkRetriesOnCollision confirms the service retries when a code already exists.
func TestCreateShortLinkRetriesOnCollision(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(4),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	repository.MarkNextCollision("ABCD")

	link, created, err := service.CreateShortLink(context.Background(), "https://example.com/path", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !created {
		t.Fatal("expected link to be newly created")
	}

	if link.Code == "" {
		t.Fatal("expected a generated code")
	}
}

// TestCreateShortLinkReusesExistingURL confirms repeat submissions return the original stored link.
func TestCreateShortLinkReusesExistingURL(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	seededLink, err := repository.CreateLink(context.Background(), "alias1", "https://example.com/reused", nil)
	if err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	link, created, err := service.CreateShortLink(context.Background(), "https://example.com/reused", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if created {
		t.Fatal("expected duplicate URL to reuse an existing link")
	}

	if link.Code != seededLink.Code {
		t.Fatalf("expected existing code %s, got %s", seededLink.Code, link.Code)
	}
}

// TestCreateShortLinkRejectsUnavailableCustomCode confirms aliases already used by another URL are blocked.
func TestCreateShortLinkRejectsUnavailableCustomCode(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(context.Background(), "team-link", "https://example.com/first", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	_, _, err := service.CreateShortLink(context.Background(), "https://example.com/second", "team-link", "")
	if !errors.Is(err, services.ErrCustomCodeUnavailable) {
		t.Fatalf("expected ErrCustomCodeUnavailable, got %v", err)
	}
}

// TestCreateShortLinkRejectsPastExpiration confirms create requests cannot use already-expired timestamps.
func TestCreateShortLinkRejectsPastExpiration(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	service := services.NewLinkServiceWithClock(
		testhelpers.NewMemoryLinkRepository(),
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)

	_, _, err := service.CreateShortLink(context.Background(), "https://example.com", "", "2026-06-27T18:00:00Z")
	if !errors.Is(err, services.ErrInvalidExpiration) {
		t.Fatalf("expected ErrInvalidExpiration, got %v", err)
	}
}

// TestResolveShortLinkRejectsExpiredLink confirms expired links no longer redirect.
func TestResolveShortLinkRejectsExpiredLink(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-time.Hour)
	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(context.Background(), "old123", "https://example.com", &expiredAt); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkServiceWithClock(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)

	_, err := service.ResolveShortLink(context.Background(), "old123")
	if !errors.Is(err, services.ErrLinkExpired) {
		t.Fatalf("expected ErrLinkExpired, got %v", err)
	}
}

// TestDeleteShortLinkMarksLinkDeleted confirms soft deletion prevents future redirects.
func TestDeleteShortLinkMarksLinkDeleted(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(context.Background(), "gone123", "https://example.com", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkServiceWithClock(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)

	if err := service.DeleteShortLink(context.Background(), "gone123"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err := service.ResolveShortLink(context.Background(), "gone123")
	if !errors.Is(err, services.ErrLinkDeleted) {
		t.Fatalf("expected ErrLinkDeleted, got %v", err)
	}
}

// TestResolveShortLinkIncrementsCount confirms redirects update click tracking state.
func TestResolveShortLinkIncrementsCount(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	_, err := repository.CreateLink(context.Background(), "abc123", "https://example.com", nil)
	if err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	originalURL, err := service.ResolveShortLink(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if originalURL != "https://example.com" {
		t.Fatalf("expected original URL to match, got %s", originalURL)
	}

	link, err := repository.GetLinkByCode(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("fetch updated link: %v", err)
	}

	if link.ClickCount != 1 {
		t.Fatalf("expected click count 1, got %d", link.ClickCount)
	}

	clickEvents, err := service.ListRecentClickEvents(context.Background(), "abc123", 10)
	if err != nil {
		t.Fatalf("list click events: %v", err)
	}

	if len(clickEvents) != 1 {
		t.Fatalf("expected 1 click event, got %d", len(clickEvents))
	}
}

// TestListRecentClickEventsReturnsNewestFirst confirms analytics events are returned newest-first.
func TestListRecentClickEventsReturnsNewestFirst(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(context.Background(), "stats1", "https://example.com", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkServiceWithClock(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)

	if err := repository.RecordClickEvent(context.Background(), "stats1", now.Add(-2*time.Minute)); err != nil {
		t.Fatalf("seed click event 1: %v", err)
	}

	if err := repository.RecordClickEvent(context.Background(), "stats1", now.Add(-time.Minute)); err != nil {
		t.Fatalf("seed click event 2: %v", err)
	}

	clickEvents, err := service.ListRecentClickEvents(context.Background(), "stats1", 2)
	if err != nil {
		t.Fatalf("list click events: %v", err)
	}

	if len(clickEvents) != 2 {
		t.Fatalf("expected 2 click events, got %d", len(clickEvents))
	}

	if !clickEvents[0].ClickedAt.After(clickEvents[1].ClickedAt) {
		t.Fatal("expected click events ordered newest first")
	}
}
