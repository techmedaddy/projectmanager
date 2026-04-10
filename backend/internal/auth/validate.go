package auth

import (
	"net/mail"
	"strings"
)

// NormalizeEmail lowercases and trims the input email before persistence and
// credential checks.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// IsValidEmail performs a small, explicit validity check suitable for request
// validation.
func IsValidEmail(email string) bool {
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	return parsed.Address == email
}
