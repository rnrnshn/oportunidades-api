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

func TestHandlerListReports(t *testing.T) {
	reportID := uuid.New()
	reporterID := uuid.New()
	entityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listReportsFn: func(context.Context, queries.ListReportsParams, ReportListFilters) ([]queries.Report, error) {
			return []queries.Report{{ID: pgtype.UUID{Bytes: [16]byte(reportID), Valid: true}, ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true}, EntityType: "course", EntityID: pgtype.UUID{Bytes: [16]byte(entityID), Valid: true}, Reason: "Informação antiga", Status: "pending"}}, nil
		},
		countReportsFn: func(context.Context, ReportListFilters) (int64, error) { return 1, nil },
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/admin/reports", handler.ListReports)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/reports", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateReportStatus(t *testing.T) {
	reportID := uuid.New()
	reporterID := uuid.New()
	reviewerID := uuid.New()
	entityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getReportByIDFn: func(context.Context, pgtype.UUID) (queries.Report, error) {
			return queries.Report{ID: pgtype.UUID{Bytes: [16]byte(reportID), Valid: true}, ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true}, EntityType: "course", EntityID: pgtype.UUID{Bytes: [16]byte(entityID), Valid: true}, Reason: "Informação antiga", Status: "pending"}, nil
		},
		updateReportStatusFn: func(context.Context, queries.UpdateReportStatusParams) (queries.Report, error) {
			return queries.Report{ID: pgtype.UUID{Bytes: [16]byte(reportID), Valid: true}, ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true}, EntityType: "course", EntityID: pgtype.UUID{Bytes: [16]byte(entityID), Valid: true}, Reason: "Informação antiga", Status: "resolved", ReviewedBy: pgtype.UUID{Bytes: [16]byte(reviewerID), Valid: true}, ModerationNotes: pgtype.Text{String: "Fixed info", Valid: true}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/admin/reports/:id", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: reviewerID.String(), Role: "admin"})
		return handler.UpdateReportStatus(c)
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/admin/reports/"+reportID.String(), strings.NewReader(`{"status":"resolved","moderation_notes":"Fixed info"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerGetReport(t *testing.T) {
	reportID := uuid.New()
	reporterID := uuid.New()
	entityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{getReportByIDFn: func(context.Context, pgtype.UUID) (queries.Report, error) {
		return queries.Report{ID: pgtype.UUID{Bytes: [16]byte(reportID), Valid: true}, ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true}, EntityType: "course", EntityID: pgtype.UUID{Bytes: [16]byte(entityID), Valid: true}, Reason: "Informação antiga", Status: "pending"}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/admin/reports/:id", handler.GetReport)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/reports/"+reportID.String(), nil)
	res, err := app.Test(req)
	if err != nil { t.Fatalf("request failed: %v", err) }
	if res.StatusCode != http.StatusOK { t.Fatalf("expected 200, got %d", res.StatusCode) }
}

func TestHandlerListReportsValidatesQuery(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/admin/reports", handler.ListReports)
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/reports?status=bad&entity_type=bad&sort=nope", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
