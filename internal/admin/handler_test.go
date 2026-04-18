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
	getArticleByIDFn     func(context.Context, pgtype.UUID) (queries.Article, error)
	publishArticleFn     func(context.Context, pgtype.UUID) (queries.Article, error)
	getOpportunityByIDFn func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	verifyOpportunityFn  func(context.Context, pgtype.UUID) (queries.Opportunity, error)
}

func (m *mockRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.getArticleByIDFn(ctx, id)
}
func (m *mockRepository) PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.publishArticleFn(ctx, id)
}
func (m *mockRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.getOpportunityByIDFn(ctx, id)
}
func (m *mockRepository) VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.verifyOpportunityFn(ctx, id)
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
