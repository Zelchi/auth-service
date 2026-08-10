package handler

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestNormalizeEmail(t *testing.T) {
	got, ok := normalizeEmail("  USER@Example.COM ")
	if !ok || got != "user@example.com" {
		t.Fatalf("normalizeEmail() = %q, %v", got, ok)
	}

	if _, ok := normalizeEmail("invalid"); ok {
		t.Fatal("normalizeEmail() accepted an invalid email")
	}
}

func TestValidPasswordUsesCharacterLimit(t *testing.T) {
	if validPassword("short") {
		t.Fatal("validPassword() accepted a short password")
	}
	valid72 := "A" + strings.Repeat("a", 70) + "1"
	if !validPassword(valid72) {
		t.Fatal("validPassword() rejected a 72-character password")
	}
	if validPassword(valid72 + "a") {
		t.Fatal("validPassword() accepted a password longer than 72 characters")
	}
	unicode72 := "Á" + strings.Repeat("á", 70) + "1"
	if !validPassword(unicode72) {
		t.Fatal("validPassword() rejected 72 Unicode characters")
	}
}

func TestValidPasswordRequiresStrengthCriteria(t *testing.T) {
	for _, password := range []string{
		"lowercase1",
		"UPPERCASE1",
		"NoNumber",
	} {
		if validPassword(password) {
			t.Fatalf("validPassword() accepted weak password %q", password)
		}
	}

	if !validPassword("StrongPassword1") {
		t.Fatal("validPassword() rejected a strong password")
	}
}

func TestPasswordHashInputSupportsUnicodeCharacterLimit(t *testing.T) {
	password := "Á" + strings.Repeat("á", 70) + "1"
	if len([]byte(password)) <= bcryptMaxInputBytes {
		t.Fatal("test password does not exceed bcrypt's byte limit")
	}
	if _, err := bcrypt.GenerateFromPassword(passwordHashInput(password), bcrypt.MinCost); err != nil {
		t.Fatalf("passwordHashInput() produced an invalid bcrypt input: %v", err)
	}
}
