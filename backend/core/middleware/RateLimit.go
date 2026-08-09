package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	startedAt time.Time
	count     int
}

const maxRateLimitEntries = 10_000

// RateLimiter é um limitador simples por processo. Ele reduz abuso em uma
// instância única; para múltiplas instâncias deve ser substituído por um
// armazenamento compartilhado.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateLimitEntry
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string]rateLimitEntry),
	}
}

func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok && len(l.entries) >= maxRateLimitEntries {
		for existingKey, existingEntry := range l.entries {
			if now.Sub(existingEntry.startedAt) >= l.window {
				delete(l.entries, existingKey)
			}
		}
		if len(l.entries) >= maxRateLimitEntries {
			return false, l.window
		}
	}
	if !ok || now.Sub(entry.startedAt) >= l.window {
		l.entries[key] = rateLimitEntry{startedAt: now, count: 1}
		return true, 0
	}

	if entry.count >= l.limit {
		return false, l.window - now.Sub(entry.startedAt)
	}

	entry.count++
	l.entries[key] = entry
	return true, 0
}

func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func requestEmail(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return ""
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(payload.Email))
}

// Email identifica tentativas pelo email informado sem confiar no valor para
// autenticação. Quando o email não está disponível, isola a chave por IP.
func Email(r *http.Request) string {
	email := requestEmail(r)
	if email == "" {
		return "missing:" + ClientIP(r)
	}
	return "email:" + email
}

func ClientIPEmail(r *http.Request) string {
	email := requestEmail(r)
	if email == "" {
		return "missing:" + ClientIP(r)
	}
	return ClientIP(r) + "|" + email
}

type RateLimitRule struct {
	Limiter *RateLimiter
	Key     func(*http.Request) string
}

func Limit(limiter *RateLimiter, key func(*http.Request) string, next http.Handler) http.Handler {
	return LimitMany([]RateLimitRule{{Limiter: limiter, Key: key}}, next)
}

func LimitMany(rules []RateLimitRule, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range rules {
			if rule.Limiter == nil || rule.Key == nil {
				continue
			}

			allowed, retryAfter := rule.Limiter.Allow(rule.Key(r))
			if !allowed {
				writeRateLimitError(w, retryAfter)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func writeRateLimitError(w http.ResponseWriter, retryAfter time.Duration) {
	seconds := int(retryAfter.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("Retry-After", formatRetryAfter(seconds))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": "muitas tentativas, tente novamente mais tarde",
	})
}

func formatRetryAfter(seconds int) string {
	return strconv.Itoa(seconds)
}
