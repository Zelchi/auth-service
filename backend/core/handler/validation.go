package handler

import (
	"net/mail"
	"strings"
)

const maxPasswordBytes = 72

func normalizeEmail(raw string) (string, bool) {
	email := strings.TrimSpace(strings.ToLower(raw))
	if len(email) == 0 || len(email) > 254 {
		return "", false
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", false
	}

	return email, true
}

func validPassword(password string) bool {
	return len([]byte(password)) >= 8 && len([]byte(password)) <= maxPasswordBytes
}

func validVerificationCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	for i := range code {
		if code[i] < '0' || code[i] > '9' {
			return false
		}
	}

	return true
}
