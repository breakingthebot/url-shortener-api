// tests/utils/validation/url_test.go
// Verifies URL normalization rules used by the link creation workflow.
// Connects to: src/utils/validation/url.go
// Created: 2026-06-28

package validation_test

import (
	"testing"

	"github.com/breakingthebot/url-shortener-api/src/utils/validation"
)

// TestNormalizeURLAcceptsHTTPS confirms a valid HTTPS URL passes validation.
func TestNormalizeURLAcceptsHTTPS(t *testing.T) {
	t.Parallel()

	normalizedURL, err := validation.NormalizeURL(" https://example.com/path ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if normalizedURL != "https://example.com/path" {
		t.Fatalf("expected normalized URL, got %s", normalizedURL)
	}
}

// TestNormalizeURLRejectsUnsupportedScheme confirms non-web schemes are blocked.
func TestNormalizeURLRejectsUnsupportedScheme(t *testing.T) {
	t.Parallel()

	if _, err := validation.NormalizeURL("ftp://example.com"); err == nil {
		t.Fatal("expected validation error for ftp URL")
	}
}
