package handler

import (
	"authentication/core/database"
	"authentication/core/email"
	"authentication/core/pending"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email e senha são obrigatórios"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "senha deve ter ao menos 8 caracteres"})
		return
	}

	var exists int
	err := database.DB.QueryRow(`SELECT 1 FROM users WHERE email = ?`, req.Email).Scan(&exists)
	if err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email já cadastrado"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao processar senha"})
		return
	}

	code, err := generateCode()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar código"})
		return
	}

	if err := email.SendVerificationCode(req.Email, code); err != nil {
		fmt.Printf("erro ao enviar email para %s: %v\n", req.Email, err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "não foi possível enviar o email de verificação, tente novamente",
		})
		return
	}

	pending.Put(req.Email, string(hash), code)

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "código enviado! verifique seu email para ativar a conta.",
	})
}
