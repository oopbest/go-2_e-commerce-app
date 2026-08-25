package domain

import "strings"

// NormalizeEmail removes surrounding whitespace and converts an email to lowercase.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
