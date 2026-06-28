package validation

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidURL       = errors.New("url must be a valid http or https URL")
	ErrInvalidShortCode = errors.New("short code must be 3-16 alphanumeric characters, hyphens, or underscores")
)

var shortCodePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,16}$`)

// ValidateURL normalizes and validates that a URL uses http or https.
func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidURL
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", ErrInvalidURL
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsed.Host == "" {
		return "", ErrInvalidURL
	}

	host := parsed.Hostname()
	if host == "" {
		return "", ErrInvalidURL
	}

	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
		return "", fmt.Errorf("%w: private or local addresses are not allowed", ErrInvalidURL)
	}

	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") {
		return "", fmt.Errorf("%w: private or local addresses are not allowed", ErrInvalidURL)
	}

	parsed.Scheme = scheme
	parsed.Fragment = ""
	return parsed.String(), nil
}

// ValidateShortCode checks custom alias format.
func ValidateShortCode(code string) error {
	if !shortCodePattern.MatchString(code) {
		return ErrInvalidShortCode
	}
	return nil
}
