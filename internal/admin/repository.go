package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("admin: not found")

type Repository interface {
	GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	ListReports(ctx context.Context, params queries.ListReportsParams) ([]queries.Report, error)
	CountReports(ctx context.Context) (int64, error)
	GetReportByID(ctx context.Context, id pgtype.UUID) (queries.Report, error)
	UpdateReportStatus(ctx context.Context, params queries.UpdateReportStatusParams) (queries.Report, error)
}
