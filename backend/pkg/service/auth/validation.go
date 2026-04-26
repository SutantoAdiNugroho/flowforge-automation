package auth

import (
	"net/mail"
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func normalizeSlug(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(strings.TrimSpace(email))
	return err == nil
}

func isValidTenantSlug(slug string) bool {
	return slugPattern.MatchString(slug)
}
