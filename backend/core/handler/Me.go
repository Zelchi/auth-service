package handler

import (
	"authentication/core/database"
	"authentication/core/middleware"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var id, email string
	var createdAt time.Time

	err := database.DB.QueryRowContext(r.Context(),
		`SELECT id, email, created_at FROM users WHERE id = ?`,
		userID,
	).Scan(&id, &email, &createdAt)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar usuário"})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "usuário não encontrado"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"email":      email,
		"created_at": createdAt,
	})
}
