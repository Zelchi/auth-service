package email

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

type providerError struct {
	statusCode int
}

const defaultResendEndpoint = "https://api.resend.com/emails"

const maxAttempts = 3

var resendHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (e *providerError) Error() string {
	return fmt.Sprintf("Resend retornou status %d", e.statusCode)
}

func SendVerificationCode(ctx context.Context, to, code string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	fromDomain := os.Getenv("RESEND_FROM")

	payload := resendRequest{
		From:    fromDomain,
		To:      []string{to},
		Subject: "Confirme sua conta",
		Html:    buildEmailHTML(code),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	endpoint := os.Getenv("RESEND_API_URL")
	if endpoint == "" {
		endpoint = defaultResendEndpoint
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("erro ao criar request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := resendHTTPClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt == maxAttempts || !isTransientNetworkError(err) {
				return fmt.Errorf("erro ao chamar API do Resend: %w", err)
			}
			if err := waitBeforeRetry(ctx, attempt); err != nil {
				return err
			}
			continue
		}

		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		_ = resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			return nil
		}
		if attempt == maxAttempts || !isTransientStatus(resp.StatusCode) {
			return &providerError{statusCode: resp.StatusCode}
		}
		if err := waitBeforeRetry(ctx, attempt); err != nil {
			return err
		}
	}

	return errors.New("falha ao enviar email")
}

func isTransientStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

func isTransientNetworkError(err error) bool {
	var networkErr net.Error
	return errors.As(err, &networkErr) && (networkErr.Timeout() || networkErr.Temporary())
}

func waitBeforeRetry(ctx context.Context, attempt int) error {
	delay := time.Duration(250*(1<<(attempt-1))) * time.Millisecond
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func buildEmailHTML(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head><meta charset="UTF-8"></head>
<body style="font-family:sans-serif;background:#f4f4f4;padding:40px">
  <div style="max-width:480px;margin:0 auto;background:#fff;border-radius:8px;padding:32px">
    <h2 style="margin-top:0;color:#111">Confirme sua conta</h2>
    <p style="color:#444">Use o código abaixo para verificar seu email. Ele expira em <strong>15 minutos</strong>.</p>
    <div style="font-size:36px;font-weight:bold;letter-spacing:8px;text-align:center;
                padding:24px;margin:24px 0;background:#f4f4f4;border-radius:6px;color:#111">
      %s
    </div>
    <p style="color:#888;font-size:13px">Se não foi você quem criou essa conta, ignore este email.</p>
  </div>
</body>
</html>`, code)
}
