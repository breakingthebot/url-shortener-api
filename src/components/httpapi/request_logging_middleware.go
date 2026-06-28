// src/components/httpapi/request_logging_middleware.go
// Assigns request IDs, enriches request context, and logs request lifecycle data consistently.
// Connects to: src/components/httpapi/request_context.go, src/components/httpapi/response_writer.go, src/main.go
// Created: 2026-06-28

package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

const requestIDHeader = "X-Request-ID"

// RequestLoggingMiddleware wraps an HTTP handler with request ID assignment and structured logging.
func RequestLoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := request.Header.Get(requestIDHeader)
		if requestID == "" {
			generatedRequestID, err := GenerateRequestID()
			if err != nil {
				logger.Error("generate request id", "error", err)
				generatedRequestID = "request-id-unavailable"
			}

			requestID = generatedRequestID
		}

		startedAt := time.Now().UTC()
		request = request.WithContext(WithRequestID(request.Context(), requestID))

		capturingWriter := NewStatusCapturingResponseWriter(writer)
		capturingWriter.Header().Set(requestIDHeader, requestID)

		logger.Info(
			"request started",
			"request_id", requestID,
			"method", request.Method,
			"path", request.URL.Path,
			"remote_addr", request.RemoteAddr,
		)

		next.ServeHTTP(capturingWriter, request)

		logger.Info(
			"request completed",
			"request_id", requestID,
			"method", request.Method,
			"path", request.URL.Path,
			"status_code", capturingWriter.StatusCode(),
			"bytes_written", capturingWriter.BytesWritten(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}
