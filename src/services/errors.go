// src/services/errors.go
// Defines service and repository sentinel errors used to map business failures to HTTP responses.
// Connects to: src/services/link_service.go, src/services/link_repository_postgres.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package services

import "errors"

var (
	// ErrInvalidURL indicates that a submitted URL failed validation.
	ErrInvalidURL = errors.New("invalid url")
	// ErrLinkNotFound indicates that a shortcode does not exist in storage.
	ErrLinkNotFound = errors.New("link not found")
	// ErrCodeCollision indicates that a generated shortcode already exists.
	ErrCodeCollision = errors.New("code collision")
)
