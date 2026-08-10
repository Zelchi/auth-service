package handler

import (
	"authentication/core/database"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func TestSessionCookieIsHttpOnly(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("COOKIE_SAMESITE", "lax")

	r := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	cookie := sessionCookie(r, "token")

	if !cookie.HttpOnly || !cookie.Secure {
		t.Fatalf("cookie flags: HttpOnly=%v Secure=%v", cookie.HttpOnly, cookie.Secure)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", cookie.SameSite)
	}
	if cookie.Path != "/" || cookie.MaxAge <= 0 {
		t.Fatalf("cookie scope: Path=%q MaxAge=%d", cookie.Path, cookie.MaxAge)
	}
}

func TestSessionCookieForCrossSiteIframeRequiresSecure(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "false")
	t.Setenv("COOKIE_SAMESITE", "none")

	r := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	cookie := sessionCookie(r, "token")

	if !cookie.Secure || cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("cookie flags: Secure=%v SameSite=%v", cookie.Secure, cookie.SameSite)
	}
}

func TestSessionCookieUsesConfiguredDomain(t *testing.T) {
	t.Setenv("COOKIE_DOMAIN", ".example.com")

	r := httptest.NewRequest(http.MethodPost, "https://auth.example.com/api/login", nil)
	cookie := sessionCookie(r, "token")

	if cookie.Domain != ".example.com" {
		t.Fatalf("Domain = %q, want %q", cookie.Domain, ".example.com")
	}
}

func TestLogoutExpiresSessionCookie(t *testing.T) {
	t.Setenv("COOKIE_SECURE", "true")
	r := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	recorder := httptest.NewRecorder()

	Logout(recorder, r)
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if cookies[0].Name != "auth-token" || cookies[0].MaxAge >= 0 || !cookies[0].HttpOnly {
		t.Fatalf("expired cookie = %+v", cookies[0])
	}
}

func TestSameOriginRequestRejectsCrossSiteRequests(t *testing.T) {
	tests := []struct {
		name   string
		origin string
		fetch  string
		ok     bool
	}{
		{name: "same origin", origin: "https://example.com", ok: true},
		{name: "different origin", origin: "https://evil.example", ok: false},
		{name: "fetch metadata", fetch: "cross-site", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "https://example.com/api/logout", nil)
			req.Header.Set("Origin", test.origin)
			if test.fetch != "" {
				req.Header.Set("Sec-Fetch-Site", test.fetch)
			}
			recorder := httptest.NewRecorder()

			if got := sameOriginRequest(recorder, req); got != test.ok {
				t.Fatalf("sameOriginRequest() = %v, want %v", got, test.ok)
			}
		})
	}
}

func TestLoginSetsSessionCookieWithoutReturningToken(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("COOKIE_SECURE", "true")

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			image TEXT NOT NULL DEFAULT '',
			name_normalized TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		t.Fatalf("create users table error = %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	if _, err := db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`, "user-123", "user@example.com", string(hash)); err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"email":"user@example.com","password":"correct-password"}`))
	recorder := httptest.NewRecorder()
	Login(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Fatalf("session cookies = %+v", cookies)
	}
	body, err := io.ReadAll(recorder.Result().Body)
	if err != nil {
		t.Fatalf("read response error = %v", err)
	}
	if strings.Contains(string(body), `"token"`) {
		t.Fatalf("login response exposed a token: %s", body)
	}
}
