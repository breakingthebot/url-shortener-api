// src/components/httpapi/request.go
// Defines small request DTOs for the JSON API layer.
// Connects to: src/components/httpapi/link_handler.go
// Created: 2026-06-28

package httpapi

// CreateLinkRequest models the incoming JSON payload for creating a short link.
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url"`
}
