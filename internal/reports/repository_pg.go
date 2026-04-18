package reports

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

func (r *PostgresRepository) CreateReport(ctx context.Context, params queries.CreateReportParams) (queries.Report, error) {
	return r.queries.CreateReport(ctx, params)
}

func (r *PostgresRepository) ReportUniversityExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.ReportUniversityExists(ctx, id)
}

func (r *PostgresRepository) ReportCourseExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.ReportCourseExists(ctx, id)
}

func (r *PostgresRepository) ReportOpportunityExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.ReportOpportunityExists(ctx, id)
}
