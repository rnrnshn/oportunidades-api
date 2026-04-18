package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct{ queries *queries.Queries }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	item, err := r.queries.GetArticleByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	item, err := r.queries.PublishArticle(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UnpublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	item, err := r.queries.UnpublishArticle(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) ArchiveArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	item, err := r.queries.ArchiveArticle(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	item, err := r.queries.GetOpportunityByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	item, err := r.queries.VerifyOpportunity(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) RejectOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	item, err := r.queries.RejectOpportunity(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) DeactivateOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	item, err := r.queries.DeactivateOpportunity(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListReports(ctx context.Context, params queries.ListReportsParams) ([]queries.Report, error) {
	return r.queries.ListReports(ctx, params)
}

func (r *PostgresRepository) CountReports(ctx context.Context) (int64, error) {
	return r.queries.CountReports(ctx)
}

func (r *PostgresRepository) GetReportByID(ctx context.Context, id pgtype.UUID) (queries.Report, error) {
	item, err := r.queries.GetReportByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Report{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateReportStatus(ctx context.Context, params queries.UpdateReportStatusParams) (queries.Report, error) {
	item, err := r.queries.UpdateReportStatus(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Report{}, ErrNotFound
	}
	return item, err
}
