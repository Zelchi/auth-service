package jwt

import (
	"strings"
	"testing"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	token, err := GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.Sub != "user-123" {
		t.Fatalf("subject = %q, want %q", claims.Sub, "user-123")
	}
}

func TestGenerateBridgeTokenIsShortLived(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	token, err := GenerateBridgeToken("user-123")
	if err != nil {
		t.Fatalf("GenerateBridgeToken() error = %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if duration := time.Duration(claims.Exp-claims.Iat) * time.Second; duration > 5*time.Minute {
		t.Fatalf("bridge token duration = %s, want at most 5m", duration)
	}
}

func TestTokenRequiresStrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	if _, err := GenerateToken("user-123"); err == nil {
		t.Fatal("GenerateToken() succeeded with a weak secret")
	}
}

func TestValidateTokenRejectsTampering(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	token, err := GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := ValidateToken(tampered); err == nil {
		t.Fatal("ValidateToken() accepted a tampered token")
	}
}

func TestValidateTokenRejectsInvalidAlgorithm(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("JWT_SECRET", secret)
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	now := time.Now()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS512, jwtv5.RegisteredClaims{
		Subject:   "user-123",
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(time.Hour)),
	})
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(signed); err == nil {
		t.Fatal("ValidateToken() accepted an unsupported algorithm")
	}
}

func TestValidateTokenRejectsInvalidType(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("JWT_SECRET", secret)
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	now := time.Now()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.RegisteredClaims{
		Subject:   "user-123",
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(time.Hour)),
	})
	token.Header["typ"] = "NOT-JWT"
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(signed); err == nil {
		t.Fatal("ValidateToken() accepted an unsupported token type")
	}
}

func TestValidateTokenRejectsExpiredToken(t *testing.T) {
	secret := strings.Repeat("s", 32)
	t.Setenv("JWT_SECRET", secret)
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	issuedAt := time.Now().Add(-2 * time.Hour)
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.RegisteredClaims{
		Subject:   "user-123",
		IssuedAt:  jwtv5.NewNumericDate(issuedAt),
		ExpiresAt: jwtv5.NewNumericDate(issuedAt.Add(time.Hour)),
	})
	token.Header["typ"] = "JWT"
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := ValidateToken(signed); err == nil {
		t.Fatal("ValidateToken() accepted an expired token")
	}
}

func TestValidateTokenAcceptsPreviousSecretDuringRotation(t *testing.T) {
	oldSecret := strings.Repeat("o", 32)
	newSecret := strings.Repeat("n", 32)
	t.Setenv("JWT_SECRET", oldSecret)
	t.Setenv("JWT_SECRET_PREVIOUS", "")

	token, err := GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	t.Setenv("JWT_SECRET", newSecret)
	t.Setenv("JWT_SECRET_PREVIOUS", oldSecret)
	if _, err := ValidateToken(token); err != nil {
		t.Fatalf("ValidateToken() rejected previous secret: %v", err)
	}

	t.Setenv("JWT_SECRET_PREVIOUS", "")
	if _, err := ValidateToken(token); err == nil {
		t.Fatal("ValidateToken() accepted token after previous secret was removed")
	}
}
