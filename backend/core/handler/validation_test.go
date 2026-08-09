package handler

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, ok := normalizeEmail("  USER@Example.COM ")
	if !ok || got != "user@example.com" {
		t.Fatalf("normalizeEmail() = %q, %v", got, ok)
	}

	if _, ok := normalizeEmail("invalid"); ok {
		t.Fatal("normalizeEmail() accepted an invalid email")
	}
}

func TestValidPasswordRejectsBcryptOverflow(t *testing.T) {
	if validPassword("short") {
		t.Fatal("validPassword() accepted a short password")
	}
	if validPassword(string(make([]byte, maxPasswordBytes+1))) {
		t.Fatal("validPassword() accepted a password longer than bcrypt's limit")
	}
}
