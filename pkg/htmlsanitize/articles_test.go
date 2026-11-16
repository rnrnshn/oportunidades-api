package htmlsanitize

import (
	"strings"
	"testing"
)

func TestArticleRemovesActiveContentAndUnsafeLinks(t *testing.T) {
	result := Article(`<h2>Safe</h2><script>alert(1)</script><img src=x onerror=alert(1)><a href="javascript:alert(1)">Bad</a><a href="https://example.com">Good</a>`)

	for _, forbidden := range []string{"<script", "<img", "javascript:"} {
		if strings.Contains(result, forbidden) {
			t.Fatalf("expected %q to be removed from %q", forbidden, result)
		}
	}
	if !strings.Contains(result, `<h2>Safe</h2>`) || !strings.Contains(result, `href="https://example.com"`) {
		t.Fatalf("expected safe formatting and link to remain, got %q", result)
	}
}

func TestArticleKeepsEditorFormatting(t *testing.T) {
	input := `<p><strong>Bold</strong> <em>italic</em> <u>underlined</u></p><blockquote><p>Quote</p></blockquote><ul><li>Item</li></ul>`
	if result := Article(input); result != input {
		t.Fatalf("expected editor formatting to remain, got %q", result)
	}
}
