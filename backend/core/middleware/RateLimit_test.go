package middleware

import (
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestRateLimiterBlocksAfterLimit(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)

	if allowed, _ := limiter.Allow("client"); !allowed {
		t.Fatal("first request was blocked")
	}
	if allowed, _ := limiter.Allow("client"); !allowed {
		t.Fatal("second request was blocked")
	}
	if allowed, retryAfter := limiter.Allow("client"); allowed || retryAfter <= 0 {
		t.Fatalf("third request allowed=%v retryAfter=%s", allowed, retryAfter)
	}
}

func TestEmailKeysPreserveRequestBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(`{"email":" USER@Example.COM "}`))
	req.RemoteAddr = "192.0.2.10:1234"

	if got := Email(req); got != "email:user@example.com" {
		t.Fatalf("Email() = %q", got)
	}
	if got := ClientIPEmail(req); got != "192.0.2.10|user@example.com" {
		t.Fatalf("ClientIPEmail() = %q", got)
	}

	if got := ClientIP(req); got != "192.0.2.10" {
		t.Fatalf("ClientIP() = %q", got)
	}
}

func TestRateLimiterCapsUniqueKeys(t *testing.T) {
	limiter := NewRateLimiter(1, time.Hour)
	for i := 0; i < maxRateLimitEntries; i++ {
		if allowed, _ := limiter.Allow("client-" + strconv.Itoa(i)); !allowed {
			t.Fatalf("key %d was blocked before capacity", i)
		}
	}
	if allowed, retryAfter := limiter.Allow("one-more"); allowed || retryAfter <= 0 {
		t.Fatalf("overflow key allowed=%v retryAfter=%s", allowed, retryAfter)
	}
}
