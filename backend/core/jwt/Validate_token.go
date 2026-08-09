package jwt

import (
	"errors"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func ValidateToken(tokenStr string) (*JWTClaims, error) {
	if tokenStr == "" {
		return nil, errors.New("token malformado")
	}

	secrets, err := signingKeys()
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, secret := range secrets {
		var claims jwtv5.RegisteredClaims
		token, parseErr := jwtv5.ParseWithClaims(
			tokenStr,
			&claims,
			func(token *jwtv5.Token) (any, error) {
				if token.Method != jwtv5.SigningMethodHS256 {
					return nil, errors.New("algoritmo de token não suportado")
				}
				if typ, ok := token.Header["typ"].(string); !ok || typ != "JWT" {
					return nil, errors.New("tipo de token não suportado")
				}
				return secret, nil
			},
			jwtv5.WithValidMethods([]string{jwtv5.SigningMethodHS256.Alg()}),
		)
		if parseErr != nil {
			lastErr = parseErr
			continue
		}
		if token == nil || !token.Valid {
			lastErr = errors.New("token inválido")
			continue
		}

		now := time.Now()
		if claims.Subject == "" {
			return nil, errors.New("token sem sujeito")
		}
		if claims.IssuedAt == nil || claims.IssuedAt.Unix() <= 0 ||
			claims.IssuedAt.Time.After(now.Add(60*time.Second)) {
			return nil, errors.New("data de emissão inválida")
		}
		if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.After(claims.IssuedAt.Time) ||
			!claims.ExpiresAt.Time.After(now) {
			return nil, errors.New("token expirado")
		}

		return &JWTClaims{
			Sub: claims.Subject,
			Iat: claims.IssuedAt.Unix(),
			Exp: claims.ExpiresAt.Unix(),
		}, nil
	}

	if lastErr == nil {
		lastErr = errors.New("token inválido")
	}
	return nil, lastErr
}
