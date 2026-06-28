// tests/utils/validation/custom_code_test.go
// Verifies optional custom shortcode validation rules for alias creation.
// Connects to: src/utils/validation/custom_code.go
// Created: 2026-06-28

package validation_test

import (
	"testing"

	"github.com/breakingthebot/url-shortener-api/src/utils/validation"
)

// TestNormalizeCustomCodeAcceptsSlug confirms a valid alias passes validation.
func TestNormalizeCustomCodeAcceptsSlug(t *testing.T) {
	t.Parallel()

	code, provided, err := validation.NormalizeCustomCode(" team_link-01 ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !provided {
		t.Fatal("expected provided to be true")
	}

	if code != "team_link-01" {
		t.Fatalf("expected trimmed code, got %s", code)
	}
}

// TestNormalizeCustomCodeRejectsInvalidCharacters confirms route-hostile aliases are blocked.
func TestNormalizeCustomCodeRejectsInvalidCharacters(t *testing.T) {
	t.Parallel()

	if _, _, err := validation.NormalizeCustomCode("bad code!"); err == nil {
		t.Fatal("expected validation error for invalid custom code")
	}
}
