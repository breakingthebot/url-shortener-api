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

// TestRedirectReturnsTemporaryRedirect confirms the redirect endpoint points to the stored URL.
func TestRedirectReturnsTemporaryRedirect(t *testing.T) {
	t.Parallel()

	repository := testhelpers.NewMemoryLinkRepository()
	if _, err := repository.CreateLink(t.Context(), "go1234", "https://example.com"); err != nil {
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
