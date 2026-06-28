// src/services/errors.go
// Defines service and repository sentinel errors used to map business failures to HTTP responses.
// Connects to: src/services/link_service.go, src/services/link_repository_postgres.go, src/components/httpapi/link_handler.go
// Created: 2026-06-28

package services

import "errors"

var (
	// ErrInvalidURL indicates that a submitted URL failed validation.
	ErrInvalidURL = errors.New("invalid url")
	// ErrInvalidCustomCode indicates that a submitted custom code failed validation.
	ErrInvalidCustomCode = errors.New("invalid custom code")
	// ErrLinkNotFound indicates that a shortcode does not exist in storage.
	ErrLinkNotFound = errors.New("link not found")
	// ErrCodeCollision indicates that a generated shortcode already exists.
	ErrCodeCollision = errors.New("code collision")
	// ErrCustomCodeUnavailable indicates that a requested custom code is already assigned elsewhere.
	ErrCustomCodeUnavailable = errors.New("custom code unavailable")
	// ErrURLAlreadyShortened indicates that a URL already exists under a different code.
	ErrURLAlreadyShortened = errors.New("url already shortened")
	// ErrInvalidExpiration indicates that a submitted expiration timestamp failed validation.
	ErrInvalidExpiration = errors.New("invalid expiration")
	// ErrLinkExpired indicates that a stored link can no longer be used because it expired.
	ErrLinkExpired = errors.New("link expired")
	// ErrLinkDeleted indicates that a stored link can no longer be used because it was soft deleted.
	ErrLinkDeleted = errors.New("link deleted")
)
