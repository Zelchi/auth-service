package handler

import (
	"authentication/core/database"
	"authentication/core/pending"
	"log/slog"
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
		methodNotAllowed(w)
		return
	}

	var req verifyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	normalizedEmail, validEmail := normalizeEmail(req.Email)
	if !validEmail {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email inválido"})
		return
	}
	req.Email = normalizedEmail
	req.Code = strings.TrimSpace(req.Code)

	if !validVerificationCode(req.Code) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email e código são obrigatórios"})
		return
	}

	reg, validCode, exists, err := pending.Check(r.Context(), req.Email, digestVerificationCode(req.Code))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar verificação"})
		return
	}
	if !exists {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "código expirado ou inexistente, registre-se novamente",
		})
		return
	}

	if !validCode {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "código inválido"})
		return
	}

	userID := uuid.New().String()
	_, err = database.DB.ExecContext(r.Context(),
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		userID, reg.Email, reg.PasswordHash,
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			_ = pending.Delete(r.Context(), req.Email)
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email já cadastrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao criar usuário"})
		return
	}

	if err := pending.Delete(r.Context(), req.Email); err != nil {
		slog.Error("pending_registration_cleanup_failed", "error", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "conta verificada com sucesso!"})
}
