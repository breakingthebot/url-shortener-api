// src/utils/validation/expiration.go
// Validates optional expiration timestamps so links can expire predictably and safely.
// Connects to: src/services/link_service.go, tests/utils/validation/expiration_test.go
// Created: 2026-06-28

package validation

import (
	"fmt"
	"strings"
	"time"
)

// NormalizeExpiration parses an optional RFC3339 timestamp and enforces that it is in the future.
func NormalizeExpiration(rawExpiration string, now time.Time) (*time.Time, error) {
	trimmedExpiration := strings.TrimSpace(rawExpiration)
	if trimmedExpiration == "" {
		return nil, nil
	}

	expiresAt, err := time.Parse(time.RFC3339, trimmedExpiration)
	if err != nil {
		return nil, fmt.Errorf("expires_at must be a valid RFC3339 timestamp: %w", err)
	}

	if !expiresAt.After(now.UTC()) {
		return nil, fmt.Errorf("expires_at must be in the future")
	}

	expiresAtUTC := expiresAt.UTC()
	return &expiresAtUTC, nil
}
