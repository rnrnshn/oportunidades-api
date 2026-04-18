package mentorship

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

func TestHandlerListMentors(t *testing.T) {
	mentorID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listMentorsFn: func(context.Context, queries.ListMentorsParams) ([]queries.ListMentorsRow, error) {
			return []queries.ListMentorsRow{{UserID: uuidToPg(mentorID), Name: "Mentor Demo", Email: "mentor@example.com", Headline: "Mentor", Bio: "Bio", Expertise: "Go"}}, nil
		},
		countMentorsFn: func(context.Context) (int64, error) { return 1, nil },
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/mentorship/mentors", handler.ListMentors)
	req := httptest.NewRequest(http.MethodGet, "/v1/mentorship/mentors", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerCreateSessionRequestRequiresAuth(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/mentorship/sessions", handler.CreateSessionRequest)
	req := httptest.NewRequest(http.MethodPost, "/v1/mentorship/sessions", strings.NewReader(`{"mentor_id":"`+uuid.NewString()+`","message":"Olá"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}

func TestHandlerCreateSessionRequestValidatesFields(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/mentorship/sessions", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
		return handler.CreateSessionRequest(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/mentorship/sessions", strings.NewReader(`{"mentor_id":"bad-id","message":"","scheduled_at":"bad-date"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerCreateSessionRequest(t *testing.T) {
	mentorID := uuid.New()
	requesterID := uuid.New()
	sessionID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getMentorByIDFn: func(context.Context, pgtype.UUID) (queries.GetMentorByIDRow, error) {
			return queries.GetMentorByIDRow{UserID: uuidToPg(mentorID), Name: "Mentor Demo", Email: "mentor@example.com", Headline: "Mentor", Bio: "Bio", Expertise: "Go"}, nil
		},
		createMentorshipSessionFn: func(context.Context, queries.CreateMentorshipSessionParams) (queries.MentorshipSession, error) {
			return queries.MentorshipSession{ID: uuidToPg(sessionID), MentorID: uuidToPg(mentorID), RequesterID: uuidToPg(requesterID), Message: "Gostaria de uma sessao", Status: "pending"}, nil
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/mentorship/sessions", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: requesterID.String(), Role: "user"})
		return handler.CreateSessionRequest(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/mentorship/sessions", strings.NewReader(`{"mentor_id":"`+mentorID.String()+`","message":"Gostaria de uma sessao"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestHandlerListSessions(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listMentorshipSessionsForUserFn: func(context.Context, queries.ListMentorshipSessionsForUserParams) ([]queries.MentorshipSession, error) {
			return []queries.MentorshipSession{{ID: uuidToPg(sessionID), MentorID: uuidToPg(uuid.New()), RequesterID: uuidToPg(userID), Message: "Oi", Status: "pending"}}, nil
		},
		countMentorshipSessionsForUserFn: func(context.Context, pgtype.UUID) (int64, error) { return 1, nil },
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/mentorship/sessions", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "user"})
		return handler.ListSessions(c)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/mentorship/sessions", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateSessionStatusValidatesPayload(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/mentorship/sessions/:id", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
		return handler.UpdateSessionStatus(c)
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/mentorship/sessions/bad-id", strings.NewReader(`{"status":"","scheduled_at":"bad-date"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
