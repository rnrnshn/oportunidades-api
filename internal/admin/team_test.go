package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
	appemail "github.com/rnrnshn/oportunidades-api/pkg/email"
)

type fakeSender struct {
	sent []appemail.InviteMessage
	err  error
}

func (f *fakeSender) SendVerificationEmail(context.Context, appemail.VerificationMessage) error {
	return nil
}
func (f *fakeSender) SendPasswordResetEmail(context.Context, appemail.PasswordResetMessage) error {
	return nil
}
func (f *fakeSender) SendInviteEmail(_ context.Context, message appemail.InviteMessage) error {
	f.sent = append(f.sent, message)
	return f.err
}

func adminApp(service *Service, actor uuid.UUID) *fiber.App {
	handler := NewHandler(service)
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	withActor := func(next fiber.Handler) fiber.Handler {
		return func(c *fiber.Ctx) error {
			c.Locals("auth_user", appauth.AuthenticatedUser{ID: actor.String(), Role: "admin"})
			return next(c)
		}
	}
	app.Post("/users/:id/suspend", withActor(handler.SuspendUser))
	app.Post("/invites", withActor(handler.CreateInvite))
	app.Get("/public/invites/:token", handler.PublicInvite)
	app.Post("/public/invites/:token/accept", handler.AcceptInvite)
	return app
}

func send(t *testing.T, app *fiber.App, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	var decoded map[string]any
	_ = json.NewDecoder(res.Body).Decode(&decoded)
	return res.StatusCode, decoded
}

func TestSuspendUser(t *testing.T) {
	actor, target := uuid.New(), uuid.New()
	active := func(context.Context, pgtype.UUID) (queries.AdminGetUserRow, error) {
		return queries.AdminGetUserRow{ID: pgUUID(target), Email: "u@example.test", Name: "U", Role: "user"}, nil
	}

	t.Run("requires a reason", func(t *testing.T) {
		status, _ := send(t, adminApp(NewService(&mockRepository{getUserFn: active}), actor), http.MethodPost, "/users/"+target.String()+"/suspend", `{"reason":"  "}`)
		if status != http.StatusUnprocessableEntity && status != http.StatusBadRequest {
			t.Fatalf("expected validation error, got %d", status)
		}
	})

	t.Run("not yourself", func(t *testing.T) {
		status, _ := send(t, adminApp(NewService(&mockRepository{getUserFn: active}), actor), http.MethodPost, "/users/"+actor.String()+"/suspend", `{"reason":"spam"}`)
		if status != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", status)
		}
	})

	t.Run("not the last admin", func(t *testing.T) {
		repo := &mockRepository{getUserFn: active, suspendUserFn: func(context.Context, queries.AdminSuspendUserParams, AuditEntry) error { return ErrLastAdmin }}
		status, _ := send(t, adminApp(NewService(repo), actor), http.MethodPost, "/users/"+target.String()+"/suspend", `{"reason":"spam"}`)
		if status != http.StatusConflict {
			t.Fatalf("expected 409, got %d", status)
		}
	})

	t.Run("already suspended", func(t *testing.T) {
		repo := &mockRepository{getUserFn: func(context.Context, pgtype.UUID) (queries.AdminGetUserRow, error) {
			return queries.AdminGetUserRow{ID: pgUUID(target), Role: "user", SuspendedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}, nil
		}}
		status, _ := send(t, adminApp(NewService(repo), actor), http.MethodPost, "/users/"+target.String()+"/suspend", `{"reason":"spam"}`)
		if status != http.StatusConflict {
			t.Fatalf("expected 409, got %d", status)
		}
	})

	t.Run("suspends with an audit entry", func(t *testing.T) {
		var gotAudit AuditEntry
		var gotParams queries.AdminSuspendUserParams
		repo := &mockRepository{getUserFn: active, suspendUserFn: func(_ context.Context, params queries.AdminSuspendUserParams, audit AuditEntry) error {
			gotParams, gotAudit = params, audit
			return nil
		}}
		status, _ := send(t, adminApp(NewService(repo), actor), http.MethodPost, "/users/"+target.String()+"/suspend", `{"reason":"  spam  "}`)
		if status != http.StatusOK {
			t.Fatalf("expected 200, got %d", status)
		}
		if gotParams.Reason != "spam" || gotParams.ActorID != pgUUID(actor) || gotAudit.Action != "suspended" || gotAudit.Target != pgUUID(target) {
			t.Fatalf("unexpected params/audit: %+v %+v", gotParams, gotAudit)
		}
	})
}

func TestCreateInvite(t *testing.T) {
	actor := uuid.New()
	body := `{"email":" New.Person@Example.test ","name":"Nova Pessoa","role":"cms_partner"}`

	t.Run("refuses an email that has an account", func(t *testing.T) {
		repo := &mockRepository{userExistsFn: func(context.Context, string) (bool, error) { return true, nil }}
		status, _ := send(t, adminApp(NewService(repo), actor), http.MethodPost, "/invites", body)
		if status != http.StatusConflict {
			t.Fatalf("expected 409, got %d", status)
		}
	})

	t.Run("refuses a second pending invite", func(t *testing.T) {
		repo := &mockRepository{pendingInviteFn: func(context.Context, string) (bool, error) { return true, nil }}
		status, _ := send(t, adminApp(NewService(repo), actor), http.MethodPost, "/invites", body)
		if status != http.StatusConflict {
			t.Fatalf("expected 409, got %d", status)
		}
	})

	t.Run("only staff roles", func(t *testing.T) {
		status, _ := send(t, adminApp(NewService(&mockRepository{}), actor), http.MethodPost, "/invites", `{"email":"a@example.test","name":"A","role":"user"}`)
		if status != http.StatusUnprocessableEntity && status != http.StatusBadRequest {
			t.Fatalf("expected validation error, got %d", status)
		}
	})

	t.Run("stores a hash, emails the link and returns it once", func(t *testing.T) {
		var stored queries.CreateUserInviteParams
		repo := &mockRepository{createInviteFn: func(_ context.Context, params queries.CreateUserInviteParams, _ AuditEntry) (queries.UserInvite, error) {
			stored = params
			return queries.UserInvite{ID: pgUUID(uuid.New()), Email: params.Email, Name: params.Name, Role: params.Role, ExpiresAt: params.ExpiresAt}, nil
		}}
		sender := &fakeSender{}
		status, decoded := send(t, adminApp(NewService(repo, WithInvites(sender, "https://dash.example.test/")), actor), http.MethodPost, "/invites", body)
		if status != http.StatusCreated {
			t.Fatalf("expected 201, got %d", status)
		}
		data := decoded["data"].(map[string]any)
		url, _ := data["invite_url"].(string)
		token := strings.TrimPrefix(url, "https://dash.example.test/convite/")
		if token == url || token == "" {
			t.Fatalf("unexpected invite url %q", url)
		}
		if stored.Email != "new.person@example.test" || stored.TokenHash != hashInviteToken(token) || stored.TokenHash == token {
			t.Fatalf("email must be normalised and only the token hash stored: %+v", stored)
		}
		if data["email_sent"] != true || len(sender.sent) != 1 || sender.sent[0].InviteURL != url {
			t.Fatalf("expected one email with the link, got %+v", sender.sent)
		}
		if d := time.Until(stored.ExpiresAt.Time); d < 6*24*time.Hour || d > 7*24*time.Hour+time.Minute {
			t.Fatalf("expected ~7 day expiry, got %v", d)
		}
	})

	t.Run("still succeeds when the email fails", func(t *testing.T) {
		repo := &mockRepository{createInviteFn: func(_ context.Context, params queries.CreateUserInviteParams, _ AuditEntry) (queries.UserInvite, error) {
			return queries.UserInvite{ID: pgUUID(uuid.New()), Email: params.Email}, nil
		}}
		status, decoded := send(t, adminApp(NewService(repo, WithInvites(&fakeSender{err: errors.New("smtp down")}, "https://d")), actor), http.MethodPost, "/invites", body)
		if status != http.StatusCreated || decoded["data"].(map[string]any)["email_sent"] != false {
			t.Fatalf("expected 201 with email_sent=false, got %d %v", status, decoded)
		}
	})
}

func TestAcceptInvite(t *testing.T) {
	token := "test-token"
	valid := queries.UserInvite{ID: pgUUID(uuid.New()), Email: "new@example.test", Name: "Nova", Role: "cms_partner", ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true}}
	byHash := func(invite queries.UserInvite) func(context.Context, string) (queries.UserInvite, error) {
		return func(_ context.Context, hash string) (queries.UserInvite, error) {
			if hash != hashInviteToken(token) {
				return queries.UserInvite{}, ErrNotFound
			}
			return invite, nil
		}
	}

	t.Run("unknown token", func(t *testing.T) {
		status, _ := send(t, adminApp(NewService(&mockRepository{inviteByHashFn: byHash(valid)}), uuid.New()), http.MethodGet, "/public/invites/other", "")
		if status != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", status)
		}
	})

	t.Run("expired", func(t *testing.T) {
		expired := valid
		expired.ExpiresAt = pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true}
		status, decoded := send(t, adminApp(NewService(&mockRepository{inviteByHashFn: byHash(expired)}), uuid.New()), http.MethodGet, "/public/invites/"+token, "")
		if status != http.StatusNotFound || !strings.Contains(decoded["error"].(map[string]any)["message"].(string), "expirou") {
			t.Fatalf("expected expired 404, got %d %v", status, decoded)
		}
	})

	t.Run("already used", func(t *testing.T) {
		used := valid
		used.AcceptedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		status, _ := send(t, adminApp(NewService(&mockRepository{inviteByHashFn: byHash(used)}), uuid.New()), http.MethodPost, "/public/invites/"+token+"/accept", `{"name":"N","password":"longenough"}`)
		if status != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", status)
		}
	})

	t.Run("short password", func(t *testing.T) {
		status, _ := send(t, adminApp(NewService(&mockRepository{inviteByHashFn: byHash(valid)}), uuid.New()), http.MethodPost, "/public/invites/"+token+"/accept", `{"name":"N","password":"short"}`)
		if status != http.StatusUnprocessableEntity && status != http.StatusBadRequest {
			t.Fatalf("expected validation error, got %d", status)
		}
	})

	t.Run("creates the account with the invited role", func(t *testing.T) {
		var created queries.AdminCreateInvitedUserParams
		repo := &mockRepository{inviteByHashFn: byHash(valid), acceptInviteFn: func(_ context.Context, _ pgtype.UUID, user queries.AdminCreateInvitedUserParams) (pgtype.UUID, error) {
			created = user
			return pgUUID(uuid.New()), nil
		}}
		status, _ := send(t, adminApp(NewService(repo), uuid.New()), http.MethodPost, "/public/invites/"+token+"/accept", `{"name":" Nova Pessoa ","password":"longenough"}`)
		if status != http.StatusCreated {
			t.Fatalf("expected 201, got %d", status)
		}
		if created.Email != valid.Email || created.Role != "cms_partner" || created.Name != "Nova Pessoa" || created.PasswordHash == "longenough" || created.PasswordHash == "" {
			t.Fatalf("unexpected account params: %+v", created)
		}
	})
}
