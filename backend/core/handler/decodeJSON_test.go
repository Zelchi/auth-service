package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsUnknownAndTrailingFields(t *testing.T) {
	tests := []string{
		`{"email":"user@example.com","unexpected":true}`,
		`{"email":"user@example.com"}{"email":"other@example.com"}`,
	}

	for _, body := range tests {
		req := httptest.NewRequest("POST", "/", strings.NewReader(body))
		var payload struct {
			Email string `json:"email"`
		}
		if err := decodeJSON(req, &payload); err == nil {
			t.Fatalf("decodeJSON() accepted invalid body %q", body)
		}
	}
}
