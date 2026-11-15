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

// AuditEntry is a change to a user account, recorded in audit_events.
type AuditEntry struct {
	ActorID pgtype.UUID
	Action  string
	Target  pgtype.UUID
	Data    map[string]any
}

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
	UserTabCounts(ctx context.Context) (queries.AdminUserTabCountsRow, error)
	CountUsersByRole(ctx context.Context) ([]queries.AdminCountUsersByRoleRow, error)
	GetUser(ctx context.Context, id pgtype.UUID) (queries.AdminGetUserRow, error)
	ListAuditEvents(ctx context.Context, userID pgtype.UUID) ([]queries.ListAuditEventsForUserRow, error)
	// UpdateUserRole changes the role, revokes the user's sessions and records the audit event in one transaction.
	UpdateUserRole(ctx context.Context, params queries.AdminUpdateUserRoleParams, audit AuditEntry) (queries.AdminUpdateUserRoleRow, error)
	// SuspendUser suspends, revokes sessions and audits in one transaction. ErrLastAdmin when blocked by the guard.
	SuspendUser(ctx context.Context, params queries.AdminSuspendUserParams, audit AuditEntry) error
	UnsuspendUser(ctx context.Context, id pgtype.UUID, audit AuditEntry) error
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	PendingInviteExists(ctx context.Context, email string) (bool, error)
	CreateInvite(ctx context.Context, params queries.CreateUserInviteParams, audit AuditEntry) (queries.UserInvite, error)
	ListPendingInvites(ctx context.Context) ([]queries.ListPendingUserInvitesRow, error)
	GetInviteByID(ctx context.Context, id pgtype.UUID) (queries.UserInvite, error)
	GetInviteByTokenHash(ctx context.Context, tokenHash string) (queries.UserInvite, error)
	RotateInvite(ctx context.Context, params queries.RotateUserInviteTokenParams, audit AuditEntry) (queries.UserInvite, error)
	RevokeInvite(ctx context.Context, id pgtype.UUID, audit AuditEntry) (queries.UserInvite, error)
	// AcceptInvite marks the invite used and creates the account in one transaction. ErrInviteInvalid when it can't be used.
	AcceptInvite(ctx context.Context, inviteID pgtype.UUID, user queries.AdminCreateInvitedUserParams) (pgtype.UUID, error)
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
