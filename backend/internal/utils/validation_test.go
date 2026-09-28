package utils_test

import (
	"testing"

	"ecommerce/backend/internal/utils"
)

func TestIsValidEmail(t *testing.T) {
	cases := map[string]bool{
		"user@example.com":    true,
		"a.b+c@sub.domain.io": true,
		"not-an-email":        false,
		"missing@domain":      false,
		"@nodomain.com":       false,
		"spaces in@email.com": false,
	}

	for email, want := range cases {
		if got := utils.IsValidEmail(email); got != want {
			t.Errorf("IsValidEmail(%q) = %v, want %v", email, got, want)
		}
	}
}

func TestIsValidPassword(t *testing.T) {
	if utils.IsValidPassword("short") {
		t.Error("expected short password to be invalid")
	}
	if !utils.IsValidPassword("longenoughpassword") {
		t.Error("expected long password to be valid")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Wireless Mouse":     "wireless-mouse",
		"  Trim Me  ":        "trim-me",
		"Special!!Chars??":   "special-chars",
		"Already-slugged":    "already-slugged",
		"":                   "item",
		"Multi   Space Name": "multi-space-name",
	}

	for input, want := range cases {
		if got := utils.Slugify(input); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", input, got, want)
		}
	}
}
