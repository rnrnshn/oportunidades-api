package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("admin: not found")

// ErrLastAdmin is returned when a role change would leave no admin.
var ErrLastAdmin = errors.New("admin: last admin")

type Repository interface {
	GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	UnpublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	ArchiveArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	RejectOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	DeactivateOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	SetUniversityVerified(ctx context.Context, params queries.SetUniversityVerifiedParams) (queries.University, error)
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
	ListUsers(ctx context.Context, params queries.AdminListUsersParams) ([]queries.AdminListUsersRow, error)
	CountUsers(ctx context.Context, params queries.AdminCountUsersParams) (int64, error)
	CountUsersByRole(ctx context.Context) ([]queries.AdminCountUsersByRoleRow, error)
	GetUser(ctx context.Context, id pgtype.UUID) (queries.AdminGetUserRow, error)
	// UpdateUserRole changes the role and revokes the user's refresh tokens in one transaction.
	UpdateUserRole(ctx context.Context, params queries.AdminUpdateUserRoleParams) (queries.AdminUpdateUserRoleRow, error)
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
