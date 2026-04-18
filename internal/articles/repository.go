package articles

import (
	"context"
	"errors"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("articles: not found")

type Repository interface {
	ListArticles(ctx context.Context, params queries.ListArticlesParams, filters Filters) ([]queries.Article, error)
	CountArticles(ctx context.Context, filters Filters) (int64, error)
	GetArticleBySlug(ctx context.Context, slug string) (queries.Article, error)
}
