package reports

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Repository interface {
	CreateReport(ctx context.Context, params queries.CreateReportParams) (queries.Report, error)
	ReportUniversityExists(ctx context.Context, id pgtype.UUID) (bool, error)
	ReportCourseExists(ctx context.Context, id pgtype.UUID) (bool, error)
	ReportOpportunityExists(ctx context.Context, id pgtype.UUID) (bool, error)
}
