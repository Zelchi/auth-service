package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func ValidateToken(tokenStr string) (*JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")

	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("token malformado")
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := b64(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errors.New("assinatura inválida")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar claims: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("erro ao parsear claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expirado")
	}

	return &claims, nil
}
