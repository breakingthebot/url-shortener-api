// src/utils/validation/url.go
// Validates and normalizes incoming URLs before they reach business logic or persistence.
// Connects to: src/services/link_service.go, tests/utils/validation/url_test.go
// Created: 2026-06-28

package validation

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeURL trims input, enforces HTTP or HTTPS, and returns a canonical string form.
func NormalizeURL(rawURL string) (string, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return "", fmt.Errorf("original_url is required")
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return "", fmt.Errorf("original_url must be a valid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("original_url must use http or https")
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("original_url must include a host")
	}

	return parsedURL.String(), nil
}
