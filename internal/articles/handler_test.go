package articles

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

func TestHandlerListArticles(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listArticlesFn: func(context.Context, queries.ListArticlesParams, Filters) ([]queries.Article, error) {
			return []queries.Article{{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "guia", Title: "Guia", Content: "Conteudo", Type: "guide"}}, nil
		},
		countArticlesFn: func(context.Context, Filters) (int64, error) { return 1, nil },
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/articles", handler.ListArticles)
	req := httptest.NewRequest(http.MethodGet, "/v1/articles?type=guide&featured=true", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerGetArticleBySlugNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{getArticleBySlugFn: func(context.Context, string) (queries.Article, error) { return queries.Article{}, ErrNotFound }}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/articles/:slug", handler.GetArticleBySlug)
	req := httptest.NewRequest(http.MethodGet, "/v1/articles/missing", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerListArticlesValidatesQuery(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/articles", handler.ListArticles)
	req := httptest.NewRequest(http.MethodGet, "/v1/articles?type=bad&featured=maybe&per_page=500", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}
