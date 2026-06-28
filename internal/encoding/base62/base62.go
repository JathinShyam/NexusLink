package base62

import (
	"errors"
	"fmt"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var (
	ErrEmptyInput = errors.New("base62: empty input")
	ErrInvalidChar = errors.New("base62: invalid character")
)

// Encode converts a non-negative integer into a base62 string.
func Encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var b strings.Builder
	for n > 0 {
		rem := n % 62
		b.WriteByte(alphabet[rem])
		n /= 62
	}

	encoded := b.String()
	runes := []byte(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Decode converts a base62 string back to an integer.
func Decode(s string) (uint64, error) {
	if s == "" {
		return 0, ErrEmptyInput
	}

	var n uint64
	for _, ch := range s {
		idx := strings.IndexRune(alphabet, ch)
		if idx < 0 {
			return 0, fmt.Errorf("%w: %q", ErrInvalidChar, ch)
		}
		n = n*62 + uint64(idx)
	}
	return n, nil
}
