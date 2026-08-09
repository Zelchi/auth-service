package handler

import (
	"authentication/core/observability"
	"authentication/core/pending"
	"log/slog"
	"net/http"
)

type resendRequest struct {
	Email string `json:"email"`
}

func Resend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	var req resendRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}

	normalizedEmail, validEmail := normalizeEmail(req.Email)
	if !validEmail {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email inválido"})
		return
	}

	code, err := generateCode()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao gerar código"})
		return
	}

	updated, err := pending.ReplaceCode(r.Context(), normalizedEmail, digestVerificationCode(code))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao preparar verificação"})
		return
	}
	if updated {
		if err := sendVerificationCode(r.Context(), normalizedEmail, code); err != nil {
			observability.IncVerificationEmailFailure()
			slog.Error("verification_email_resend_failed", "error", err)
		}
	}

	// A mesma resposta é usada quando não existe cadastro pendente para não
	// revelar quais emails já iniciaram o fluxo.
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "se houver um cadastro pendente, um novo código será enviado",
	})
}
