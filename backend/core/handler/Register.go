package handler

import (
	"authentication/core/database"
	"authentication/core/email"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var err error
	var hash []byte

	if r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	hash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao processar senha"})
		return
	}

	userID := uuid.New().String()

	_, err = database.DB.Exec(
		`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`,
		userID, req.Email, string(hash),
	)

	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email já cadastrado"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar usuário"})
		return
	}

	code, err := generateCode()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar código"})
		return
	}

	codeID := uuid.New().String()
	expiresAt := time.Now().UTC().Add(15 * time.Minute)

	if _, err = database.DB.Exec(
		`INSERT INTO verification_codes (id, user_id, code, expires_at) VALUES (?, ?, ?, ?)`,
		codeID, userID, code, expiresAt,
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao salvar código"})
		return
	}

	if err := email.SendVerificationCode(req.Email, code); err != nil {
		fmt.Printf("aviso: erro ao enviar email para %s: %v\n", req.Email, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao enviar email de verificação"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "conta criada! verifique seu email para ativar.",
		"user_id": userID,
	})
}
