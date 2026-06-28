// src/utils/shortcode/generator.go
// Generates collision-resistant shortcode strings for shortened links using cryptographic randomness.
// Connects to: src/services/link_service.go
// Created: 2026-06-28

package shortcode

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Generator creates random short codes of a fixed length.
type Generator struct {
	length int
}

// NewGenerator constructs a shortcode generator for the requested length.
func NewGenerator(length int) Generator {
	return Generator{length: length}
}

// Generate returns a new random shortcode suitable for use as a URL slug.
func (g Generator) Generate() (string, error) {
	buffer := make([]byte, g.length)
	maxIndex := big.NewInt(int64(len(alphabet)))

	for index := range buffer {
		randomIndex, err := rand.Int(rand.Reader, maxIndex)
		if err != nil {
			return "", fmt.Errorf("generate shortcode: %w", err)
		}

		buffer[index] = alphabet[randomIndex.Int64()]
	}

	return string(buffer), nil
}
