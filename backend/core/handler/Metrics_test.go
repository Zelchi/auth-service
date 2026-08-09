package handler

import (
	"authentication/core/observability"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsExposesVerificationEmailFailures(t *testing.T) {
	before := observability.VerificationEmailFailures()
	observability.IncVerificationEmailFailure()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()
	Metrics(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), "auth_verification_email_failures_total "+fmt.Sprint(before+1)) {
		t.Fatalf("metrics body does not contain updated counter: %s", recorder.Body.String())
	}
}
