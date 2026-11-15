package admin

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

func pgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte(id), Valid: true} }

func roleApp(repo *mockRepository, actorID uuid.UUID) *fiber.App {
	handler := NewHandler(NewService(repo))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/admin/users/:id/role", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: actorID.String(), Role: "admin"})
		return handler.UpdateUserRole(c)
	})
	return app
}

func patchRole(t *testing.T, app *fiber.App, userID uuid.UUID, body string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodPatch, "/v1/admin/users/"+userID.String()+"/role", strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return res.StatusCode
}

func TestHandlerListUsers(t *testing.T) {
	var got queries.AdminListUsersParams
	handler := NewHandler(NewService(&mockRepository{
		listUsersFn: func(_ context.Context, params queries.AdminListUsersParams) ([]queries.AdminListUsersRow, error) {
			got = params
			return []queries.AdminListUsersRow{{ID: pgUUID(uuid.New()), Email: "ana@example.com", Name: "Ana", Role: "user"}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/admin/users", handler.ListUsers)

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1/admin/users?role=mentor&q=ana&page=3&per_page=10&sort=name_asc", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if got.Role.String != "mentor" || got.Query.String != "ana" || got.Sort != "name_asc" || got.Limit != 10 || got.Offset != 20 {
		t.Fatalf("unexpected list params: %+v", got)
	}

	res, _ = app.Test(httptest.NewRequest(http.MethodGet, "/v1/admin/users?role=superuser", nil))
	if res.StatusCode != http.StatusUnprocessableEntity && res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected validation error for bad role, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateUserRole(t *testing.T) {
	actorID := uuid.New()
	userID := uuid.New()
	existing := func(role string) func(context.Context, pgtype.UUID) (queries.AdminGetUserRow, error) {
		return func(context.Context, pgtype.UUID) (queries.AdminGetUserRow, error) {
			return queries.AdminGetUserRow{ID: pgUUID(userID), Email: "u@example.com", Name: "U", Role: role}, nil
		}
	}

	t.Run("promotes a user", func(t *testing.T) {
		var updated queries.AdminUpdateUserRoleParams
		app := roleApp(&mockRepository{
			getUserFn: existing("user"),
			updateUserRoleFn: func(_ context.Context, params queries.AdminUpdateUserRoleParams, _ AuditEntry) (queries.AdminUpdateUserRoleRow, error) {
				updated = params
				return queries.AdminUpdateUserRoleRow{ID: params.ID, Email: "u@example.com", Name: "U", Role: params.Role}, nil
			},
		}, actorID)
		if status := patchRole(t, app, userID, `{"role":"cms_partner"}`); status != http.StatusOK {
			t.Fatalf("expected 200, got %d", status)
		}
		if updated.Role != "cms_partner" || updated.ID != pgUUID(userID) {
			t.Fatalf("unexpected update params: %+v", updated)
		}
	})

	t.Run("rejects own role change", func(t *testing.T) {
		called := false
		app := roleApp(&mockRepository{
			getUserFn: existing("admin"),
			updateUserRoleFn: func(context.Context, queries.AdminUpdateUserRoleParams, AuditEntry) (queries.AdminUpdateUserRoleRow, error) {
				called = true
				return queries.AdminUpdateUserRoleRow{}, nil
			},
		}, actorID)
		if status := patchRole(t, app, actorID, `{"role":"user"}`); status != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", status)
		}
		if called {
			t.Fatal("update must not run for own role change")
		}
	})

	t.Run("keeps the last admin", func(t *testing.T) {
		app := roleApp(&mockRepository{
			getUserFn: existing("admin"),
			updateUserRoleFn: func(context.Context, queries.AdminUpdateUserRoleParams, AuditEntry) (queries.AdminUpdateUserRoleRow, error) {
				return queries.AdminUpdateUserRoleRow{}, ErrLastAdmin
			},
		}, actorID)
		if status := patchRole(t, app, userID, `{"role":"user"}`); status != http.StatusConflict {
			t.Fatalf("expected 409, got %d", status)
		}
	})

	t.Run("rejects unknown role", func(t *testing.T) {
		app := roleApp(&mockRepository{getUserFn: existing("user")}, actorID)
		status := patchRole(t, app, userID, `{"role":"owner"}`)
		if status != http.StatusUnprocessableEntity && status != http.StatusBadRequest {
			t.Fatalf("expected validation error, got %d", status)
		}
	})

	t.Run("unknown user", func(t *testing.T) {
		app := roleApp(&mockRepository{}, actorID)
		if status := patchRole(t, app, userID, `{"role":"mentor"}`); status != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", status)
		}
	})
}
