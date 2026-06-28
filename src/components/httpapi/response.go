// src/components/httpapi/response.go
// Provides consistent JSON response helpers for success and error API payloads.
// Connects to: src/components/httpapi/link_handler.go
// Created: 2026-06-28

package httpapi

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse models a JSON error payload returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON sends a JSON response with the provided status code and payload.
func WriteJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

// WriteError sends a standardized JSON error payload.
func WriteError(writer http.ResponseWriter, statusCode int, message string) {
	WriteJSON(writer, statusCode, ErrorResponse{Error: message})
}
