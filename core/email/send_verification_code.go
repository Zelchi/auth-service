package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

type resendError struct {
	Name       string `json:"name"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
}

func SendVerificationCode(to, code string) error {
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

	req, err := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("erro ao criar request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro ao chamar API do Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		var resendErr resendError
		if err := json.Unmarshal(raw, &resendErr); err == nil && resendErr.Message != "" {
			return fmt.Errorf("resend error %d: %s", resendErr.StatusCode, resendErr.Message)
		}
		return fmt.Errorf("resend retornou status %d: %s", resp.StatusCode, string(raw))
	}

	return nil
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
