package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	getArticleByIDFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	publishArticleFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	unpublishArticleFn      func(context.Context, pgtype.UUID) (queries.Article, error)
	archiveArticleFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	getOpportunityByIDFn    func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	verifyOpportunityFn     func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	rejectOpportunityFn     func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	deactivateOpportunityFn func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	listReportsFn           func(context.Context, queries.ListReportsParams, ReportListFilters) ([]queries.Report, error)
	countReportsFn          func(context.Context, ReportListFilters) (int64, error)
	getReportByIDFn         func(context.Context, pgtype.UUID) (queries.Report, error)
	updateReportStatusFn    func(context.Context, queries.UpdateReportStatusParams) (queries.Report, error)
}

func (m *mockRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.getArticleByIDFn(ctx, id)
}
func (m *mockRepository) PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.publishArticleFn(ctx, id)
}
func (m *mockRepository) UnpublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.unpublishArticleFn(ctx, id)
}
func (m *mockRepository) ArchiveArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.archiveArticleFn(ctx, id)
}
func (m *mockRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.getOpportunityByIDFn(ctx, id)
}
func (m *mockRepository) VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.verifyOpportunityFn(ctx, id)
}
func (m *mockRepository) RejectOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.rejectOpportunityFn(ctx, id)
}
func (m *mockRepository) DeactivateOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.deactivateOpportunityFn(ctx, id)
}
func (m *mockRepository) ListReports(ctx context.Context, params queries.ListReportsParams, filters ReportListFilters) ([]queries.Report, error) {
	return m.listReportsFn(ctx, params, filters)
}
func (m *mockRepository) CountReports(ctx context.Context, filters ReportListFilters) (int64, error) {
	return m.countReportsFn(ctx, filters)
}
func (m *mockRepository) GetReportByID(ctx context.Context, id pgtype.UUID) (queries.Report, error) {
	return m.getReportByIDFn(ctx, id)
}
func (m *mockRepository) UpdateReportStatus(ctx context.Context, params queries.UpdateReportStatusParams) (queries.Report, error) {
	return m.updateReportStatusFn(ctx, params)
}

func TestHandlerPublishArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
		publishArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "published", PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/publish", handler.PublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/publish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerVerifyOpportunityNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
		return queries.Opportunity{}, ErrNotFound
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/verify", handler.VerifyOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+uuid.NewString()+"/verify", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerPublishArticleValidatesID(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/publish", handler.PublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/bad-id/publish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerUnpublishArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "published"}, nil
		},
		unpublishArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/unpublish", handler.UnpublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/unpublish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerArchiveArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
		archiveArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "archived"}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/archive", handler.ArchiveArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/archive", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerRejectOpportunity(t *testing.T) {
	opportunityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: true}, nil
		},
		rejectOpportunityFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: false, IsActive: false}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/reject", handler.RejectOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+opportunityID.String()+"/reject", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerDeactivateOpportunity(t *testing.T) {
	opportunityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: true}, nil
		},
		deactivateOpportunityFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: false}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/deactivate", handler.DeactivateOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+opportunityID.String()+"/deactivate", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}
