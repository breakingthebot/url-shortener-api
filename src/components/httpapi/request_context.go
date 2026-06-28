// src/components/httpapi/request_context.go
// Stores and retrieves request-scoped values used by HTTP middleware and handlers.
// Connects to: src/components/httpapi/request_logging_middleware.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package httpapi

import "context"

type requestContextKey string

const requestIDContextKey requestContextKey = "request_id"

// WithRequestID stores a request ID in the request context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

// RequestIDFromContext returns the stored request ID when one exists.
func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}

	return requestID
}
