package handler

import (
	"crypto/sha256"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	minPasswordCharacters = 8
	maxPasswordCharacters = 72
	bcryptMaxInputBytes   = 72
)

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
	if !validPasswordLength(password) {
		return false
	}

	var hasLower, hasUpper, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasNumber = true
		}
	}

	return hasLower && hasUpper && hasNumber
}

func validPasswordLength(password string) bool {
	characters := utf8.RuneCountInString(password)
	return characters >= minPasswordCharacters && characters <= maxPasswordCharacters
}

// passwordHashInput preserves the bcrypt format for existing ASCII passwords
// and pre-hashes longer UTF-8 input so 72 characters remain a valid limit.
func passwordHashInput(password string) []byte {
	input := []byte(password)
	if len(input) <= bcryptMaxInputBytes {
		return input
	}

	digest := sha256.Sum256(input)
	return digest[:]
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
