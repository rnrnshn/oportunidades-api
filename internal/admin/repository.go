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
	UnpublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	ArchiveArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	RejectOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	DeactivateOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	ListReports(ctx context.Context, params queries.ListReportsParams, filters ReportListFilters) ([]queries.Report, error)
	CountReports(ctx context.Context, filters ReportListFilters) (int64, error)
	GetReportByID(ctx context.Context, id pgtype.UUID) (queries.Report, error)
	UpdateReportStatus(ctx context.Context, params queries.UpdateReportStatusParams) (queries.Report, error)
	CountCMSArticles(ctx context.Context) (int64, error)
	CountCMSOpportunities(ctx context.Context) (int64, error)
	CountPendingReports(ctx context.Context) (int64, error)
	CountPendingMentorshipSessions(ctx context.Context) (int64, error)
	ListRecentPendingReports(ctx context.Context) ([]queries.Report, error)
	ListRecentUnverifiedOpportunities(ctx context.Context) ([]queries.Opportunity, error)
	ContentCreatedPerDay(ctx context.Context, days int) ([]DailyContentCount, error)
	OpportunitiesByType(ctx context.Context) ([]TypeCount, error)
	OpportunitiesByStatus(ctx context.Context) ([]TypeCount, error)
}

type DailyContentCount struct {
	Date          string `json:"date"`
	Articles      int64  `json:"articles"`
	Opportunities int64  `json:"opportunities"`
}

type TypeCount struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}
