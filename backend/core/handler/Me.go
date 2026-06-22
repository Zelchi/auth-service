package handler

import (
	"authentication/core/database"
	"authentication/core/middleware"
	"net/http"
	"time"
)

func Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var id, email string
	var createdAt time.Time

	err := database.DB.QueryRow(
		`SELECT id, email, created_at FROM users WHERE id = ?`,
		userID,
	).Scan(&id, &email, &createdAt)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "usuário não encontrado"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":         id,
		"email":      email,
		"created_at": createdAt,
	})
}
