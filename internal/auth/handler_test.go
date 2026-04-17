package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

func TestHandlerRegisterValidatesRequiredFields(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/auth/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", strings.NewReader(`{"email":"","password":"","name":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerLoginValidatesRequiredFields(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/auth/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"email":"","password":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerRefreshRequiresCookie(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}, Config{
		JWTSecret:          "secret",
		JWTExpiry:          15 * time.Minute,
		RefreshTokenExpiry: 30 * 24 * time.Hour,
		RefreshCookieName:  "refresh_token",
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/auth/refresh", handler.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}
