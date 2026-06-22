package handler

import (
	"authentication/core/database"
	"authentication/core/jwt"
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var userID, hash string
	err := database.DB.QueryRow(
		`SELECT id, password FROM users WHERE email = ?`,
		req.Email,
	).Scan(&userID, &hash)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciais inválidas"})
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciais inválidas"})
		return
	}

	token, err := jwt.GenerateToken(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar token"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":   token,
		"user_id": userID,
	})
}
