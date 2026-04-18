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
	createArticleFn      func(context.Context, queries.CreateArticleParams) (queries.Article, error)
	createOpportunityFn  func(context.Context, queries.CreateOpportunityParams) (queries.Opportunity, error)
	getArticleByIDFn     func(context.Context, pgtype.UUID) (queries.Article, error)
	updateArticleFn      func(context.Context, queries.UpdateArticleParams) (queries.Article, error)
	getOpportunityByIDFn func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	updateOpportunityFn  func(context.Context, queries.UpdateOpportunityParams) (queries.Opportunity, error)
}

func (m *mockRepository) CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error) {
	return m.createArticleFn(ctx, params)
}
func (m *mockRepository) CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error) {
	return m.createOpportunityFn(ctx, params)
}
func (m *mockRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.getArticleByIDFn(ctx, id)
}
func (m *mockRepository) UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error) {
	return m.updateArticleFn(ctx, params)
}
func (m *mockRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.getOpportunityByIDFn(ctx, id)
}
func (m *mockRepository) UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error) {
	return m.updateOpportunityFn(ctx, params)
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

func TestHandlerUpdateArticle(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Novo Artigo", Content: "Conteudo", Type: "guide", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
		},
		updateArticleFn: func(context.Context, queries.UpdateArticleParams) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Artigo Editado", Content: "Conteudo editado", Type: "news", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/cms/articles/:id", handler.UpdateArticle)
	req := httptest.NewRequest(http.MethodPatch, "/v1/cms/articles/"+articleID.String(), strings.NewReader(`{"title":"Artigo Editado","content":"Conteudo editado","type":"news"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateOpportunityValidatesID(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/cms/opportunities/:id", handler.UpdateOpportunity)
	req := httptest.NewRequest(http.MethodPatch, "/v1/cms/opportunities/bad-id", strings.NewReader(`{"title":"Teste"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
