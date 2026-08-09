package handler

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const sessionDuration = 24 * time.Hour

func sessionCookie(r *http.Request, token string) *http.Cookie {
	secure := cookieSecure(r)
	sameSite := cookieSameSite()
	if sameSite == http.SameSiteNoneMode {
		secure = true
	}

	return &http.Cookie{
		Name:     "auth-token",
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionDuration / time.Second),
		Expires:  time.Now().Add(sessionDuration),
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}

func expiredSessionCookie(r *http.Request) *http.Cookie {
	cookie := sessionCookie(r, "")
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0)
	return cookie
}

func cookieSecure(r *http.Request) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE"))) {
	case "1", "true", "yes":
		return true
	case "0", "false", "no":
		return false
	default:
		return r.TLS != nil
	}
}

func cookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SAMESITE"))) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func sameOriginRequest(w http.ResponseWriter, r *http.Request) bool {
	if strings.EqualFold(strings.TrimSpace(r.Header.Get("Sec-Fetch-Site")), "cross-site") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "origem não permitida"})
		return false
	}

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		!strings.EqualFold(parsed.Host, r.Host) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "origem não permitida"})
		return false
	}

	return true
}
