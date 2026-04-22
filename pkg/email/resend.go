package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const resendSendURL = "https://api.resend.com/emails"

type VerificationMessage struct {
	To              string
	RecipientName   string
	VerificationURL string
}

type Sender interface {
	SendVerificationEmail(ctx context.Context, message VerificationMessage) error
}

type ResendSender struct {
	apiKey     string
	from       string
	httpClient *http.Client
}

func NewResendSender(apiKey string, from string) *ResendSender {
	return &ResendSender{
		apiKey: strings.TrimSpace(apiKey),
		from:   strings.TrimSpace(from),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ResendSender) SendVerificationEmail(ctx context.Context, message VerificationMessage) error {
	if s == nil || s.apiKey == "" {
		return fmt.Errorf("email: resend api key is not configured")
	}
	if s.from == "" {
		return fmt.Errorf("email: from address is not configured")
	}
	if strings.TrimSpace(message.To) == "" {
		return fmt.Errorf("email: recipient is required")
	}
	if strings.TrimSpace(message.VerificationURL) == "" {
		return fmt.Errorf("email: verification url is required")
	}

	recipientName := strings.TrimSpace(message.RecipientName)
	if recipientName == "" {
		recipientName = ""
	}

	payload := map[string]any{
		"from":    s.from,
		"to":      []string{strings.TrimSpace(message.To)},
		"subject": "Verifica o teu email no Oportunidades",
		"html":    verificationHTML(recipientName, message.VerificationURL),
		"text":    verificationText(recipientName, message.VerificationURL),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("email: marshal resend payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, resendSendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: build resend request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("email: send resend request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email: resend returned status %d", resp.StatusCode)
	}

	return nil
}

func verificationHTML(recipientName string, verificationURL string) string {
	greeting := "Ola,"
	if recipientName != "" {
		greeting = fmt.Sprintf("Ola %s,", recipientName)
	}

	return fmt.Sprintf(`<div style="font-family: Arial, sans-serif; line-height: 1.5; color: #111827;">
	<p>%s</p>
	<p>Confirma o teu email para activares a tua conta no Oportunidades.</p>
	<p><a href="%s">Verificar email</a></p>
	<p>Se nao criaste esta conta, podes ignorar esta mensagem.</p>
	</div>`, greeting, verificationURL)
}

func verificationText(recipientName string, verificationURL string) string {
	greeting := "Ola,"
	if recipientName != "" {
		greeting = fmt.Sprintf("Ola %s,", recipientName)
	}

	return fmt.Sprintf("%s\n\nConfirma o teu email para activares a tua conta no Oportunidades.\n\nVerificar email: %s\n\nSe nao criaste esta conta, podes ignorar esta mensagem.\n", greeting, verificationURL)
}
