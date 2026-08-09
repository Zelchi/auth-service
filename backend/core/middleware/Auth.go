package middleware

import (
	"authentication/core/jwt"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var tokenStr string
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}

		if tokenStr == "" {
			if cookie, err := r.Cookie("auth-token"); err == nil {
				tokenStr = strings.TrimSpace(cookie.Value)
			}
		}

		if tokenStr == "" {
			writeAuthError(w, http.StatusUnauthorized, "token não informado")
			return
		}

		claims, err := jwt.ValidateToken(tokenStr)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "token inválido")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.Sub)
		next(w, r.WithContext(ctx))
	}
}
