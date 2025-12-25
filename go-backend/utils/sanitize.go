package utils

import (
	"html"
	"regexp"
	"strings"
)

// SanitizeString removes dangerous characters and limits length
func SanitizeString(input string, maxLength int) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Escape HTML entities to prevent XSS
	sanitized = html.EscapeString(sanitized)

	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	// Limit length
	if maxLength > 0 && len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return sanitized
}

// ValidateUsername checks if username is safe (alphanumeric, underscore, hyphen)
func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	return matched
}

// SanitizeDescription sanitizes text descriptions
func SanitizeDescription(description string) string {
	return SanitizeString(description, 500)
}

// SanitizeCategoryName sanitizes category names
func SanitizeCategoryName(name string) string {
	return SanitizeString(name, 100)
}
