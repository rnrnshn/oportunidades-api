package reports

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

type mockRepository struct {
	createReportFn func(context.Context, queries.CreateReportParams) (queries.Report, error)
}

func (m *mockRepository) CreateReport(ctx context.Context, params queries.CreateReportParams) (queries.Report, error) {
	return m.createReportFn(ctx, params)
}

func TestHandlerCreateReport(t *testing.T) {
	reportID := uuid.New()
	reporterID := uuid.New()
	entityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{createReportFn: func(context.Context, queries.CreateReportParams) (queries.Report, error) {
		return queries.Report{ID: pgtype.UUID{Bytes: [16]byte(reportID), Valid: true}, ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true}, EntityType: "course", EntityID: pgtype.UUID{Bytes: [16]byte(entityID), Valid: true}, Reason: "Informação antiga", Status: "pending"}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/reports", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: reporterID.String(), Role: "user"})
		return handler.Create(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/reports", strings.NewReader(`{"entity_type":"course","entity_id":"`+entityID.String()+`","reason":"Informação antiga"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestHandlerCreateReportValidatesFields(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/reports", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
		return handler.Create(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/reports", strings.NewReader(`{"entity_type":"","entity_id":"bad-id","reason":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
