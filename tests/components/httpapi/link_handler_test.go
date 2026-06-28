// tests/components/httpapi/link_handler_test.go
// Verifies the API layer returns the expected HTTP status codes and payloads.
// Connects to: src/components/httpapi/link_handler.go, tests/testhelpers/memory_link_repository.go
// Created: 2026-06-28

package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/breakingthebot/url-shortener-api/src/components/httpapi"
	"github.com/breakingthebot/url-shortener-api/src/services"
	"github.com/breakingthebot/url-shortener-api/src/utils/shortcode"
	"github.com/breakingthebot/url-shortener-api/tests/testhelpers"
)

// TestCreateLinkReturnsCreated confirms the create endpoint accepts a valid payload.
func TestCreateLinkReturnsCreated(t *testing.T) {
	t.Parallel()

	service := services.NewLinkService(
		testhelpers.NewMemoryLinkRepository(),
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body, err := json.Marshal(httpapi.CreateLinkRequest{OriginalURL: "https://example.com"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

// TestCreateLinkReturnsOKForDuplicateURL confirms duplicate URLs reuse the existing link.
func TestCreateLinkReturnsOKForDuplicateURL(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "saved1", "https://example.com/existing", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body, err := json.Marshal(httpapi.CreateLinkRequest{OriginalURL: "https://example.com/existing"})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

// TestCreateLinkReturnsConflictForUnavailableCustomCode confirms alias collisions surface as conflicts.
func TestCreateLinkReturnsConflictForUnavailableCustomCode(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "launch", "https://example.com/first", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body, err := json.Marshal(httpapi.CreateLinkRequest{
		OriginalURL: "https://example.com/second",
		CustomCode:  "launch",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/links", bytes.NewReader(body))
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", recorder.Code)
	}
}

// TestRedirectReturnsTemporaryRedirect confirms the redirect endpoint points to the stored URL.
func TestRedirectReturnsTemporaryRedirect(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "go1234", "https://example.com", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/go1234", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", recorder.Code)
	}

	if location := recorder.Header().Get("Location"); location != "https://example.com" {
		t.Fatalf("expected redirect location, got %s", location)
	}
}

// TestRedirectReturnsGoneForExpiredLink confirms expired links no longer redirect.
func TestRedirectReturnsGoneForExpiredLink(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	expiredAt := now.Add(-time.Hour)
	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "expired1", "https://example.com", &expiredAt); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkServiceWithClock(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/expired1", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGone {
		t.Fatalf("expected 410, got %d", recorder.Code)
	}
}

// TestDeleteLinkReturnsNoContent confirms soft delete endpoint marks a link unavailable.
func TestDeleteLinkReturnsNoContent(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "delete1", "https://example.com", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkService(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodDelete, "/api/v1/links/delete1", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}

// TestGetClickEventsReturnsRecentHistory confirms analytics endpoint returns recent redirect events.
func TestGetClickEventsReturnsRecentHistory(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "stats2", "https://example.com", nil); err != nil {
		t.Fatalf("seed repository: %v", err)
	}

	service := services.NewLinkServiceWithClock(
		repository,
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func() time.Time { return now },
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	for range 2 {
		redirectRequest := httptest.NewRequest(http.MethodGet, "/stats2", nil)
		redirectRecorder := httptest.NewRecorder()
		mux.ServeHTTP(redirectRecorder, redirectRequest)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/links/stats2/clicks?limit=2", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var clickEvents []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &clickEvents); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(clickEvents) != 2 {
		t.Fatalf("expected 2 click events, got %d", len(clickEvents))
	}
}

// TestGetClickEventsRejectsInvalidLimit confirms analytics limit validation is enforced.
func TestGetClickEventsRejectsInvalidLimit(t *testing.T) {
	t.Parallel()

	service := services.NewLinkService(
		testhelpers.NewMemoryLinkRepository(),
		shortcode.NewGenerator(6),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	handler := httpapi.NewLinkHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/links/missing/clicks?limit=abc", nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
