package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
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

type PasswordResetMessage struct {
	To            string
	RecipientName string
	ResetURL      string
}

type InviteMessage struct {
	To            string
	RecipientName string
	InviterName   string
	RoleLabel     string
	InviteURL     string
}

type Sender interface {
	SendVerificationEmail(ctx context.Context, message VerificationMessage) error
	SendPasswordResetEmail(ctx context.Context, message PasswordResetMessage) error
	SendInviteEmail(ctx context.Context, message InviteMessage) error
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
	if strings.TrimSpace(message.VerificationURL) == "" {
		return fmt.Errorf("email: verification url is required")
	}
	content := actionEmail{
		Greeting:   greeting(message.RecipientName),
		Intro:      "Confirma o teu email para activares a tua conta no Oportunidades.",
		ButtonText: "Verificar email",
		URL:        message.VerificationURL,
		Outro:      "Se não criaste esta conta, podes ignorar esta mensagem.",
	}
	return s.send(ctx, message.To, "Verifica o teu email no Oportunidades", content)
}

func (s *ResendSender) SendPasswordResetEmail(ctx context.Context, message PasswordResetMessage) error {
	if strings.TrimSpace(message.ResetURL) == "" {
		return fmt.Errorf("email: reset url is required")
	}
	content := actionEmail{
		Greeting:   greeting(message.RecipientName),
		Intro:      "Recebemos um pedido para redefinir a password da tua conta no Oportunidades.",
		ButtonText: "Redefinir password",
		URL:        message.ResetURL,
		Outro:      "Se não fizeste este pedido, ignora esta mensagem. A tua password continua igual.",
	}
	return s.send(ctx, message.To, "Redefine a tua password no Oportunidades", content)
}

func (s *ResendSender) SendInviteEmail(ctx context.Context, message InviteMessage) error {
	if strings.TrimSpace(message.InviteURL) == "" {
		return fmt.Errorf("email: invite url is required")
	}
	content := actionEmail{
		Greeting:   greeting(message.RecipientName),
		Intro:      fmt.Sprintf("%s convidou-te para a equipa do Oportunidades como %s. Aceita o convite para criares a tua conta no painel.", message.InviterName, message.RoleLabel),
		ButtonText: "Aceitar convite",
		URL:        message.InviteURL,
		Outro:      "O convite expira em 7 dias. Se não esperavas esta mensagem, podes ignorá-la.",
	}
	return s.send(ctx, message.To, "Convite para a equipa do Oportunidades", content)
}

func (s *ResendSender) send(ctx context.Context, to string, subject string, content actionEmail) error {
	if s == nil || s.apiKey == "" {
		return fmt.Errorf("email: resend api key is not configured")
	}
	if s.from == "" {
		return fmt.Errorf("email: from address is not configured")
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return fmt.Errorf("email: recipient is required")
	}

	body, err := json.Marshal(map[string]any{
		"from":    s.from,
		"to":      []string{to},
		"subject": subject,
		"html":    content.HTML(),
		"text":    content.Text(),
	})
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

type actionEmail struct {
	Greeting   string
	Intro      string
	ButtonText string
	URL        string
	Outro      string
}

func greeting(recipientName string) string {
	name := strings.TrimSpace(recipientName)
	if name == "" {
		return "Olá,"
	}
	return fmt.Sprintf("Olá %s,", name)
}

func (e actionEmail) HTML() string {
	return fmt.Sprintf(`<div style="font-family: Arial, sans-serif; line-height: 1.5; color: #111827;">
	<p>%s</p>
	<p>%s</p>
	<p><a href="%s">%s</a></p>
	<p>%s</p>
	</div>`,
		html.EscapeString(e.Greeting),
		html.EscapeString(e.Intro),
		html.EscapeString(e.URL),
		html.EscapeString(e.ButtonText),
		html.EscapeString(e.Outro),
	)
}

func (e actionEmail) Text() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s: %s\n\n%s\n", e.Greeting, e.Intro, e.ButtonText, e.URL, e.Outro)
}
