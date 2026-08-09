package handler

import "net/http"

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if !sameOriginRequest(w, r) {
		return
	}

	http.SetCookie(w, expiredSessionCookie(r))
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"message": "sessão encerrada"})
}
