package handler

import (
	"authentication/core/database"
	"authentication/core/email"
	"authentication/core/observability"
	"authentication/core/pending"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

var sendVerificationCode = email.SendVerificationCode

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	normalizedEmail, validEmail := normalizeEmail(req.Email)
	if !validEmail || req.Password == "" || req.PasswordConfirmation == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, senha e confirmação de senha são obrigatórios e válidos"})
		return
	}
	req.Email = normalizedEmail
	if !validPassword(req.Password) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "senha deve ter entre 8 e 72 caracteres e conter maiúscula, minúscula e número"})
		return
	}
	if req.Password != req.PasswordConfirmation {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "as senhas não coincidem"})
		return
	}

	var exists int
	err := database.DB.QueryRowContext(r.Context(), `SELECT 1 FROM users WHERE email = ?`, req.Email).Scan(&exists)
	if err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email já cadastrado"})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar usuário"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword(passwordHashInput(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao processar senha"})
		return
	}

	code, err := generateCode()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar código"})
		return
	}

	if err := pending.Put(r.Context(), req.Email, string(hash), digestVerificationCode(code)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao preparar verificação"})
		return
	}

	if err := sendVerificationCode(r.Context(), req.Email, code); err != nil {
		observability.IncVerificationEmailFailure()
		slog.Error("verification_email_failed", "error", err)
		_ = pending.Delete(r.Context(), req.Email)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "não foi possível enviar o email de verificação, tente novamente",
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "código enviado! verifique seu email para ativar a conta.",
	})
}
