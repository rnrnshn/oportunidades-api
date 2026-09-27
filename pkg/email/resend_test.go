package email

import (
	"strings"
	"testing"
)

func TestActionEmailEscapesUserInput(t *testing.T) {
	content := actionEmail{
		Greeting:   greeting(`<a href="https://evil.example">Clica</a>`),
		Intro:      "Intro",
		ButtonText: "Verificar email",
		URL:        "https://oportunidades.social/auth/verify-email?token=a&b=c",
		Outro:      "Outro",
	}

	got := content.HTML()
	if strings.Contains(got, `<a href="https://evil.example">`) {
		t.Fatalf("recipient name was not escaped: %s", got)
	}
	if !strings.Contains(got, `href="https://oportunidades.social/auth/verify-email?token=a&amp;b=c"`) {
		t.Fatalf("expected escaped action url, got: %s", got)
	}
}

func TestGreeting(t *testing.T) {
	if got := greeting("  "); got != "Olá," {
		t.Fatalf("unexpected greeting for empty name: %q", got)
	}
	if got := greeting("Ana"); got != "Olá Ana," {
		t.Fatalf("unexpected greeting: %q", got)
	}
}
