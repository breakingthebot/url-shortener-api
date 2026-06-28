// src/models/link.go
// Defines the core link domain model used by services, handlers, and storage.
// Connects to: src/services/link_service.go, src/services/link_repository_postgres.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package models

import "time"

// Link represents a shortened URL and its persisted analytics fields.
type Link struct {
	Code        string    `json:"code"`
	OriginalURL string    `json:"original_url"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}
