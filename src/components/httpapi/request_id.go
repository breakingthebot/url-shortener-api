// src/components/httpapi/request_id.go
// Generates request identifiers for HTTP tracing and response correlation.
// Connects to: src/components/httpapi/request_logging_middleware.go
// Created: 2026-06-28

package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const requestIDByteLength = 16

// GenerateRequestID creates a random request identifier suitable for headers and logs.
func GenerateRequestID() (string, error) {
	buffer := make([]byte, requestIDByteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate request id: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}
