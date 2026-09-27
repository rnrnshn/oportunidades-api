package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

func TestRateLimitPerIPAndPath(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler, ProxyHeader: "CF-Connecting-IP"})
	limit := RateLimit(2, time.Minute)
	ok := func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }
	app.Post("/login", limit, ok)
	app.Post("/register", limit, ok)

	send := func(path, ip string) *http.Response {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("CF-Connecting-IP", ip)
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		return res
	}

	for i := 0; i < 2; i++ {
		if res := send("/login", "1.1.1.1"); res.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, res.StatusCode)
		}
	}

	res := send("/login", "1.1.1.1")
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res.StatusCode)
	}
	var body apierror.Response
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error.Code != "RATE_LIMITED" {
		t.Fatalf("expected RATE_LIMITED, got %+v", body)
	}

	if res := send("/login", "2.2.2.2"); res.StatusCode != http.StatusOK {
		t.Fatalf("other IP: expected 200, got %d", res.StatusCode)
	}
	if res := send("/register", "1.1.1.1"); res.StatusCode != http.StatusOK {
		t.Fatalf("other path: expected 200, got %d", res.StatusCode)
	}
}
