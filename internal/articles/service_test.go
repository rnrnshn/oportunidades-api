package articles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	listArticlesFn     func(context.Context, queries.ListArticlesParams, Filters) ([]queries.Article, error)
	countArticlesFn    func(context.Context, Filters) (int64, error)
	getArticleBySlugFn func(context.Context, string) (queries.Article, error)
}

func (m *mockRepository) ListArticles(ctx context.Context, params queries.ListArticlesParams, filters Filters) ([]queries.Article, error) {
	return m.listArticlesFn(ctx, params, filters)
}
func (m *mockRepository) CountArticles(ctx context.Context, filters Filters) (int64, error) {
	return m.countArticlesFn(ctx, filters)
}
func (m *mockRepository) GetArticleBySlug(ctx context.Context, slug string) (queries.Article, error) {
	return m.getArticleBySlugFn(ctx, slug)
}

func TestListArticles(t *testing.T) {
	articleID := uuid.New()
	service := NewService(&mockRepository{
		listArticlesFn: func(context.Context, queries.ListArticlesParams, Filters) ([]queries.Article, error) {
			return []queries.Article{{ID: uuidToPg(articleID), Slug: "guia", Title: "Guia", Content: "Conteudo", Type: "guide", IsFeatured: true, PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}}, nil
		},
		countArticlesFn: func(context.Context, Filters) (int64, error) { return 1, nil },
	})
	result, err := service.ListArticles(context.Background(), PaginationParams{}, Filters{})
	if err != nil {
		t.Fatalf("list articles returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Meta.Total != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetArticleBySlugNotFound(t *testing.T) {
	service := NewService(&mockRepository{getArticleBySlugFn: func(context.Context, string) (queries.Article, error) { return queries.Article{}, ErrNotFound }})
	_, err := service.GetArticleBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func uuidToPg(value uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte(value), Valid: true} }
