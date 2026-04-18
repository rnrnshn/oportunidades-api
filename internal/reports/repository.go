package reports

import (
	"context"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Repository interface {
	CreateReport(ctx context.Context, params queries.CreateReportParams) (queries.Report, error)
}
