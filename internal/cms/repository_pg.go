package cms

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct{ queries *queries.Queries }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error) {
	return r.queries.CreateArticle(ctx, params)
}

func (r *PostgresRepository) CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error) {
	return r.queries.CreateOpportunity(ctx, params)
}
