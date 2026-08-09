package email

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func setEmailTestClient(t *testing.T, status int, body string, check func(*http.Request)) {
	t.Helper()
	previousClient := resendHTTPClient
	resendHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if check != nil {
			check(r)
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}
	t.Cleanup(func() { resendHTTPClient = previousClient })
	t.Setenv("RESEND_API_URL", "https://resend.test/emails")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "noreply@example.com")
}

func TestSendVerificationCodeDoesNotReturnProviderBody(t *testing.T) {
	setEmailTestClient(t, http.StatusBadGateway, `{"message":"provider-secret-response"}`, nil)

	err := SendVerificationCode(context.Background(), "user@example.com", "123456")
	if err == nil {
		t.Fatal("SendVerificationCode() returned nil for provider failure")
	}
	if strings.Contains(err.Error(), "provider-secret-response") {
		t.Fatalf("provider response leaked: %v", err)
	}
	if !strings.Contains(err.Error(), "502") {
		t.Fatalf("provider status missing from error: %v", err)
	}
}

func TestSendVerificationCodeAcceptsCreatedResponse(t *testing.T) {
	setEmailTestClient(t, http.StatusCreated, "", func(r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Errorf("Authorization header = %q", r.Header.Get("Authorization"))
		}
	})

	if err := SendVerificationCode(context.Background(), "user@example.com", "123456"); err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}
}

func TestSendVerificationCodeRetriesTransientResponses(t *testing.T) {
	var attempts atomic.Int32
	previousClient := resendHTTPClient
	resendHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		current := attempts.Add(1)
		status := http.StatusServiceUnavailable
		if current == 2 {
			status = http.StatusCreated
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}
	t.Cleanup(func() { resendHTTPClient = previousClient })
	t.Setenv("RESEND_API_URL", "https://resend.test/emails")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "noreply@example.com")

	if err := SendVerificationCode(context.Background(), "user@example.com", "123456"); err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

func TestSendVerificationCodeDoesNotRetryPermanentResponses(t *testing.T) {
	var attempts atomic.Int32
	previousClient := resendHTTPClient
	resendHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts.Add(1)
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})}
	t.Cleanup(func() { resendHTTPClient = previousClient })
	t.Setenv("RESEND_API_URL", "https://resend.test/emails")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "noreply@example.com")

	if err := SendVerificationCode(context.Background(), "user@example.com", "123456"); err == nil {
		t.Fatal("SendVerificationCode() returned nil for permanent failure")
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("attempts = %d, want 1", got)
	}
}
