// tests/utils/validation/expiration_test.go
// Verifies optional expiration timestamp validation for expiring short links.
// Connects to: src/utils/validation/expiration.go
// Created: 2026-06-28

package validation_test

import (
	"testing"
	"time"

	"github.com/breakingthebot/url-shortener-api/src/utils/validation"
)

// TestNormalizeExpirationAcceptsFutureTimestamp confirms a future RFC3339 timestamp is accepted.
func TestNormalizeExpirationAcceptsFutureTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	expiresAt, err := validation.NormalizeExpiration("2026-06-29T18:00:00Z", now)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if expiresAt == nil || expiresAt.Format(time.RFC3339) != "2026-06-29T18:00:00Z" {
		t.Fatalf("expected parsed expiration, got %v", expiresAt)
	}
}

// TestNormalizeExpirationRejectsPastTimestamp confirms expired timestamps are blocked.
func TestNormalizeExpirationRejectsPastTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 28, 18, 0, 0, 0, time.UTC)
	if _, err := validation.NormalizeExpiration("2026-06-27T18:00:00Z", now); err == nil {
		t.Fatal("expected validation error for past expiration")
	}
}
