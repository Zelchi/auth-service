package handler

import (
	"authentication/core/database"
	"authentication/core/pending"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type verifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Code = strings.TrimSpace(req.Code)

	if req.Email == "" || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email e código são obrigatórios"})
		return
	}

	reg, ok := pending.Get(req.Email)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "código expirado ou inexistente, registre-se novamente",
		})
		return
	}

	if reg.Code != req.Code {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "código inválido"})
		return
	}

	userID := uuid.New().String()
	_, err := database.DB.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		userID, reg.Email, reg.PasswordHash,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			pending.Delete(req.Email)
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email já cadastrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao criar usuário"})
		return
	}

	pending.Delete(req.Email)

	writeJSON(w, http.StatusOK, map[string]string{"message": "conta verificada com sucesso!"})
}
