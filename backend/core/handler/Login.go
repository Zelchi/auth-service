package handler

import (
	"authentication/core/database"
	"authentication/core/jwt"
	"database/sql"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	normalizedEmail, validEmail := normalizeEmail(req.Email)
	if !validEmail || !validPasswordLength(req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciais inválidas"})
		return
	}
	req.Email = normalizedEmail

	var userID, hash string
	err := database.DB.QueryRowContext(r.Context(),
		`SELECT id, password FROM users WHERE email = ?`,
		req.Email,
	).Scan(&userID, &hash)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar usuário"})
			return
		}
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciais inválidas"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hash), passwordHashInput(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciais inválidas"})
		return
	}

	token, err := jwt.GenerateToken(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar token"})
		return
	}

	http.SetCookie(w, sessionCookie(r, token))
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{
		"user_id": userID,
	})
}
