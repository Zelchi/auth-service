package middleware

import (
	"authentication/core/jwt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthAcceptsHttpOnlySessionCookie(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	token, err := jwt.GenerateToken("user-123")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.AddCookie(&http.Cookie{Name: "auth-token", Value: token})
	recorder := httptest.NewRecorder()

	Auth(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(UserIDKey); got != "user-123" {
			t.Errorf("user id = %v, want user-123", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}
