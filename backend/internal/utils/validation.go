package utils

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail does a pragmatic (not fully RFC 5322) email format check.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPassword requires at least 8 characters. Kept simple and
// documented so it is easy to tighten for production use.
func IsValidPassword(password string) bool {
	return len(password) >= 8
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a name into a URL-friendly slug, e.g.
// "Wireless Mouse!" -> "wireless-mouse".
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugNonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "item"
	}
	return s
}
