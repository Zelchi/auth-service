package handler

import (
	"authentication/core/database"
	"encoding/json"
	"net/http"
	"strings"
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

	var userID string
	err := database.DB.QueryRow(
		`SELECT id FROM users WHERE email = ? AND verified = 0`,
		req.Email,
	).Scan(&userID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "usuário não encontrado ou já verificado"})
		return
	}

	var codeID string
	err = database.DB.QueryRow(
		`SELECT id FROM verification_codes
		 WHERE user_id = ? AND code = ? AND used = 0 AND expires_at > datetime('now')`,
		userID, req.Code,
	).Scan(&codeID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "código inválido ou expirado"})
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro interno"})
		return
	}
	defer tx.Rollback()

	if _, err = tx.Exec(`UPDATE users SET verified = 1 WHERE id = ?`, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao verificar conta"})
		return
	}
	if _, err = tx.Exec(`UPDATE verification_codes SET used = 1 WHERE id = ?`, codeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao invalidar código"})
		return
	}

	if err = tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao confirmar operação"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "conta verificada com sucesso!"})
}
