package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"os"
	"time"
)

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTClaims struct {
	Sub string `json:"sub"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

func GenerateToken(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	header, _ := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	claims, _ := json.Marshal(JWTClaims{
		Sub: userID,
		Iat: time.Now().Unix(),
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	})

	payload := b64(header) + "." + b64(claims)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := b64(mac.Sum(nil))

	return payload + "." + sig, nil
}
