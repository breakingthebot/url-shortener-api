// src/utils/validation/custom_code.go
// Validates optional custom short codes so aliases stay readable, safe, and route-friendly.
// Connects to: src/services/link_service.go, tests/utils/validation/custom_code_test.go
// Created: 2026-06-28

package validation

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	minCustomCodeLength = 4
	maxCustomCodeLength = 32
)

var customCodePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// NormalizeCustomCode trims and validates an optional custom shortcode.
func NormalizeCustomCode(rawCode string) (string, bool, error) {
	trimmedCode := strings.TrimSpace(rawCode)
	if trimmedCode == "" {
		return "", false, nil
	}

	if len(trimmedCode) < minCustomCodeLength || len(trimmedCode) > maxCustomCodeLength {
		return "", false, fmt.Errorf("custom_code must be between %d and %d characters", minCustomCodeLength, maxCustomCodeLength)
	}

	if !customCodePattern.MatchString(trimmedCode) {
		return "", false, fmt.Errorf("custom_code may only contain letters, numbers, hyphens, and underscores")
	}

	return trimmedCode, true, nil
}
