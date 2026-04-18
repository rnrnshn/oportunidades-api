package reports

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct{ queries *queries.Queries }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) CreateReport(ctx context.Context, params queries.CreateReportParams) (queries.Report, error) {
	return r.queries.CreateReport(ctx, params)
}
