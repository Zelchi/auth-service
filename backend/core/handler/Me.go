package handler

import (
	"authentication/core/database"
	"authentication/core/middleware"
	models "authentication/core/model"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxProfileNameRunes  = 80
	maxProfileImageBytes = 512 * 1024
)

type profileUpdateRequest struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}

func Me(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getMe(w, r)
	case http.MethodPatch:
		UpdateMe(w, r)
	default:
		methodNotAllowed(w)
	}
}

func getMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
		return
	}

	user, err := findUser(r.Context(), userID)
	if err != nil {
		writeUserLookupError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, user)
}

func UpdateMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		methodNotAllowed(w)
		return
	}
	if !sameOriginRequest(w, r) {
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "sessão inválida"})
		return
	}

	var req profileUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "body inválido"})
		return
	}
	if req.Name == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "nome é obrigatório"})
		return
	}

	name, normalizedName, err := normalizeProfileName(*req.Name)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	current, err := findUser(r.Context(), userID)
	if err != nil {
		writeUserLookupError(w, err)
		return
	}

	image := current.Image
	if req.Image != nil {
		image = strings.TrimSpace(*req.Image)
		if err := validateProfileImage(image); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	var conflictingID string
	err = database.DB.QueryRowContext(r.Context(),
		`SELECT id FROM users WHERE name_normalized = ? AND id <> ?`,
		normalizedName, userID,
	).Scan(&conflictingID)
	if err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "esse nome já está em uso"})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar nome"})
		return
	}

	_, err = database.DB.ExecContext(r.Context(),
		`UPDATE users SET name = ?, name_normalized = ?, image = ? WHERE id = ?`,
		name, normalizedName, image, userID,
	)
	if err != nil {
		if isUniqueConstraint(err) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "esse nome já está em uso"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao atualizar perfil"})
		return
	}

	updated, err := findUser(r.Context(), userID)
	if err != nil {
		writeUserLookupError(w, err)
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, updated)
}

func findUser(ctx context.Context, userID string) (models.User, error) {
	var user models.User
	err := database.DB.QueryRowContext(ctx,
		`SELECT id, email, name, image, created_at FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Image, &user.CreatedAt)
	return user, err
}

func writeUserLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "usuário não encontrado"})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erro ao consultar usuário"})
}

func normalizeProfileName(value string) (string, string, error) {
	if !utf8.ValidString(value) {
		return "", "", errors.New("nome inválido")
	}

	name := strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if name == "" {
		return "", "", errors.New("nome é obrigatório")
	}
	if utf8.RuneCountInString(name) > maxProfileNameRunes {
		return "", "", errors.New("nome deve ter no máximo 80 caracteres")
	}
	for _, char := range name {
		if unicode.IsControl(char) {
			return "", "", errors.New("nome contém caracteres inválidos")
		}
	}

	return name, strings.ToLower(name), nil
}

func validateProfileImage(value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 900_000 {
		return errors.New("foto muito grande; use uma imagem de até 512 KB")
	}

	if strings.HasPrefix(value, "data:") {
		separator := strings.IndexByte(value, ',')
		if separator < 0 || !strings.HasSuffix(value[:separator], ";base64") {
			return errors.New("foto inválida; use PNG, JPEG, WEBP ou GIF")
		}

		mediaType := strings.TrimPrefix(value[:separator], "data:")
		switch strings.ToLower(strings.TrimSuffix(mediaType, ";base64")) {
		case "image/png", "image/jpeg", "image/webp", "image/gif":
		default:
			return errors.New("foto inválida; use PNG, JPEG, WEBP ou GIF")
		}

		decoded, err := base64.StdEncoding.DecodeString(value[separator+1:])
		if err != nil || len(decoded) == 0 || len(decoded) > maxProfileImageBytes {
			return errors.New("foto muito grande ou inválida; use uma imagem de até 512 KB")
		}
		return nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("foto inválida; envie uma imagem ou uma URL HTTP/HTTPS")
	}
	return nil
}

func isUniqueConstraint(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
