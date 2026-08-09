package jwt

import (
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func GenerateToken(userID string) (string, error) {
	return generateToken(userID, 24*time.Hour)
}

// GenerateBridgeToken cria uma credencial curta para integrações autorizadas.
// Ela não prolonga a sessão principal, que continua protegida pelo cookie.
func GenerateBridgeToken(userID string) (string, error) {
	return generateToken(userID, 5*time.Minute)
}

func generateToken(userID string, duration time.Duration) (string, error) {
	secret, err := signingKey()
	if err != nil {
		return "", err
	}

	now := time.Now()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, jwtv5.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwtv5.NewNumericDate(now),
		ExpiresAt: jwtv5.NewNumericDate(now.Add(duration)),
	})
	token.Header["typ"] = "JWT"

	return token.SignedString(secret)
}
