package cms

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
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

func (r *PostgresRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return r.queries.GetArticleByID(ctx, id)
}

func (r *PostgresRepository) UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error) {
	return r.queries.UpdateArticle(ctx, params)
}

func (r *PostgresRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return r.queries.GetOpportunityByID(ctx, id)
}

func (r *PostgresRepository) UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error) {
	return r.queries.UpdateOpportunity(ctx, params)
}
