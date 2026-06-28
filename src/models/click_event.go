// src/models/click_event.go
// Defines the click event model used for recent analytics and event history responses.
// Connects to: src/services/link_service.go, src/services/link_repository_postgres.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package models

import "time"

// ClickEvent represents a single successful redirect event for a short link.
type ClickEvent struct {
	Code      string    `json:"code"`
	ClickedAt time.Time `json:"clicked_at"`
}
