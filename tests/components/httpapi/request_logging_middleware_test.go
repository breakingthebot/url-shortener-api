// tests/components/httpapi/request_logging_middleware_test.go
// Verifies request logging middleware behavior around request IDs and response headers.
// Connects to: src/components/httpapi/request_logging_middleware.go, src/components/httpapi/request_context.go
// Created: 2026-06-28

package httpapi_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/breakingthebot/url-shortener-api/src/components/httpapi"
)

// TestRequestLoggingMiddlewareEchoesIncomingRequestID confirms provided request IDs are preserved.
func TestRequestLoggingMiddlewareEchoesIncomingRequestID(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := httpapi.RequestLoggingMiddleware(logger, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if requestID := httpapi.RequestIDFromContext(request.Context()); requestID != "req-123" {
			t.Fatalf("expected request id req-123, got %s", requestID)
		}

		writer.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Request-ID", "req-123")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Request-ID") != "req-123" {
		t.Fatalf("expected response request id header, got %s", recorder.Header().Get("X-Request-ID"))
	}
}

// TestRequestLoggingMiddlewareGeneratesRequestID confirms requests without an ID receive one.
func TestRequestLoggingMiddlewareGeneratesRequestID(t *testing.T) {
	t.Parallel()

	var observedRequestID string
	logBuffer := bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	handler := httpapi.RequestLoggingMiddleware(logger, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		observedRequestID = httpapi.RequestIDFromContext(request.Context())
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if observedRequestID == "" {
		t.Fatal("expected generated request id in context")
	}

	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected generated request id header")
	}
}
