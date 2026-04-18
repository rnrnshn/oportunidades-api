package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

func TestHandlerGetMe(t *testing.T) {
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{getUserByIDFn: func(context.Context, pgtype.UUID) (queries.User, error) {
		return queries.User{ID: uuidToPg(userID), Email: "user@example.com", Role: "user", Name: "User"}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/account/me", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "user"})
		return handler.GetMe(c)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/account/me", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateMeValidatesName(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/account/me", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
		return handler.UpdateMe(c)
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/account/me", strings.NewReader(`{"name":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerChangePasswordValidatesPayload(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/account/password", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
		return handler.ChangePassword(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/account/password", strings.NewReader(`{"current_password":"","new_password":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
