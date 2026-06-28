// src/components/httpapi/response_writer.go
// Wraps the HTTP response writer so middleware can inspect status codes and response sizes.
// Connects to: src/components/httpapi/request_logging_middleware.go
// Created: 2026-06-28

package httpapi

import "net/http"

// StatusCapturingResponseWriter records response metadata for logging middleware.
type StatusCapturingResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// NewStatusCapturingResponseWriter wraps an HTTP response writer with status capture support.
func NewStatusCapturingResponseWriter(writer http.ResponseWriter) *StatusCapturingResponseWriter {
	return &StatusCapturingResponseWriter{
		ResponseWriter: writer,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader stores and forwards the HTTP status code.
func (w *StatusCapturingResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write records bytes written while preserving the response status behavior.
func (w *StatusCapturingResponseWriter) Write(payload []byte) (int, error) {
	bytesWritten, err := w.ResponseWriter.Write(payload)
	w.bytesWritten += bytesWritten
	return bytesWritten, err
}

// StatusCode returns the final observed HTTP status code.
func (w *StatusCapturingResponseWriter) StatusCode() int {
	return w.statusCode
}

// BytesWritten returns the number of bytes written to the client.
func (w *StatusCapturingResponseWriter) BytesWritten() int {
	return w.bytesWritten
}
