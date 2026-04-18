package cms

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
	createArticleFn     func(context.Context, queries.CreateArticleParams) (queries.Article, error)
	createOpportunityFn func(context.Context, queries.CreateOpportunityParams) (queries.Opportunity, error)
}

func (m *mockRepository) CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error) {
	return m.createArticleFn(ctx, params)
}
func (m *mockRepository) CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error) {
	return m.createOpportunityFn(ctx, params)
}

func TestHandlerCreateArticle(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{createArticleFn: func(context.Context, queries.CreateArticleParams) (queries.Article, error) {
		return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Novo Artigo", Type: "guide", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/articles", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateArticle(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/articles", strings.NewReader(`{"title":"Novo Artigo","content":"Conteudo","type":"guide"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestHandlerCreateOpportunityValidation(t *testing.T) {
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/opportunities", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateOpportunity(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/opportunities", strings.NewReader(`{"title":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
