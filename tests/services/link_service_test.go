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

	_, err := service.CreateShortLink(context.Background(), "not-a-url")
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

	link, err := service.CreateShortLink(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if link.Code == "" {
		t.Fatal("expected a generated code")
	}
}

// TestResolveShortLinkIncrementsCount confirms redirects update click tracking state.
func TestResolveShortLinkIncrementsCount(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	_, err := repository.CreateLink(context.Background(), "abc123", "https://example.com")
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
}
