package admin

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct {
	queries *queries.Queries
	pool    *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool), pool: pool}
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

func (r *PostgresRepository) SetUniversityVerified(ctx context.Context, params queries.SetUniversityVerifiedParams) (queries.University, error) {
	item, err := r.queries.SetUniversityVerified(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.University{}, ErrNotFound
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

func (r *PostgresRepository) ListReports(ctx context.Context, params queries.ListReportsParams, filters ReportListFilters) ([]queries.Report, error) {
	return r.queries.ListReports(ctx, params)
}

func (r *PostgresRepository) CountReports(ctx context.Context, filters ReportListFilters) (int64, error) {
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

func (r *PostgresRepository) CountCMSArticles(ctx context.Context) (int64, error) {
	return r.queries.CountCMSArticles(ctx)
}

func (r *PostgresRepository) CountCMSOpportunities(ctx context.Context) (int64, error) {
	return r.queries.CountCMSOpportunities(ctx)
}

func (r *PostgresRepository) CountPendingReports(ctx context.Context) (int64, error) {
	row := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM reports WHERE status = 'pending' AND deleted_at IS NULL`)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (r *PostgresRepository) CountPendingMentorshipSessions(ctx context.Context) (int64, error) {
	row := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM mentorship_sessions WHERE status = 'pending' AND deleted_at IS NULL`)
	var count int64
	err := row.Scan(&count)
	return count, err
}

func (r *PostgresRepository) ListRecentPendingReports(ctx context.Context) ([]queries.Report, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, reporter_id, entity_type, entity_id, reason, status, reviewed_by, moderation_notes, resolved_at, created_at, updated_at, deleted_at FROM reports WHERE status = 'pending' AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []queries.Report{}
	for rows.Next() {
		var i queries.Report
		if err := rows.Scan(
			&i.ID,
			&i.ReporterID,
			&i.EntityType,
			&i.EntityID,
			&i.Reason,
			&i.Status,
			&i.ReviewedBy,
			&i.ModerationNotes,
			&i.ResolvedAt,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListRecentUnverifiedOpportunities(ctx context.Context) ([]queries.Opportunity, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, slug, title, type, entity_name, description, requirements, deadline, apply_url, external_url_label, country, location, is_remote, language, area, hero_image_url, provider_logo_url, amount_min, amount_max, amount_currency, coverage, eligibility, application_process, degree_level, program_area, is_active, published_by, verified, created_at, updated_at, deleted_at FROM opportunities WHERE verified = FALSE AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 5`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []queries.Opportunity{}
	for rows.Next() {
		var i queries.Opportunity
		if err := rows.Scan(
			&i.ID,
			&i.Slug,
			&i.Title,
			&i.Type,
			&i.EntityName,
			&i.Description,
			&i.Requirements,
			&i.Deadline,
			&i.ApplyUrl,
			&i.ExternalUrlLabel,
			&i.Country,
			&i.Location,
			&i.IsRemote,
			&i.Language,
			&i.Area,
			&i.HeroImageUrl,
			&i.ProviderLogoUrl,
			&i.AmountMin,
			&i.AmountMax,
			&i.AmountCurrency,
			&i.Coverage,
			&i.Eligibility,
			&i.ApplicationProcess,
			&i.DegreeLevel,
			&i.ProgramArea,
			&i.IsActive,
			&i.PublishedBy,
			&i.Verified,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ContentCreatedPerDay(ctx context.Context, days int) ([]DailyContentCount, error) {
	query := `
		WITH dates AS (
			SELECT generate_series(
				(CURRENT_DATE - make_interval(days => $1::int))::date,
				CURRENT_DATE,
				'1 day'::interval
			)::date AS day
		),
		art AS (
			SELECT created_at::date AS day, COUNT(*) AS cnt
			FROM articles WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE - make_interval(days => $1::int)
			GROUP BY 1
		),
		opp AS (
			SELECT created_at::date AS day, COUNT(*) AS cnt
			FROM opportunities WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE - make_interval(days => $1::int)
			GROUP BY 1
		)
		SELECT d.day::text, COALESCE(a.cnt, 0), COALESCE(o.cnt, 0)
		FROM dates d
		LEFT JOIN art a ON a.day = d.day
		LEFT JOIN opp o ON o.day = d.day
		ORDER BY d.day`
	rows, err := r.pool.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []DailyContentCount
	for rows.Next() {
		var item DailyContentCount
		if err := rows.Scan(&item.Date, &item.Articles, &item.Opportunities); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) OpportunitiesByType(ctx context.Context) ([]TypeCount, error) {
	rows, err := r.pool.Query(ctx, `SELECT type, COUNT(*) FROM opportunities WHERE deleted_at IS NULL GROUP BY type ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TypeCount
	for rows.Next() {
		var item TypeCount
		if err := rows.Scan(&item.Label, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) OpportunitiesByStatus(ctx context.Context) ([]TypeCount, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			CASE
				WHEN verified = TRUE AND is_active = TRUE THEN 'active'
				WHEN verified = TRUE AND is_active = FALSE THEN 'inactive'
				WHEN verified = FALSE THEN 'pending'
			END AS status,
			COUNT(*)
		FROM opportunities WHERE deleted_at IS NULL
		GROUP BY 1 ORDER BY COUNT(*) DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []TypeCount
	for rows.Next() {
		var item TypeCount
		if err := rows.Scan(&item.Label, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListUsers(ctx context.Context, params queries.AdminListUsersParams) ([]queries.AdminListUsersRow, error) {
	return r.queries.AdminListUsers(ctx, params)
}

func (r *PostgresRepository) CountUsers(ctx context.Context, params queries.AdminCountUsersParams) (int64, error) {
	return r.queries.AdminCountUsers(ctx, params)
}

func (r *PostgresRepository) UserTabCounts(ctx context.Context) (queries.AdminUserTabCountsRow, error) {
	return r.queries.AdminUserTabCounts(ctx)
}

func (r *PostgresRepository) CountUsersByRole(ctx context.Context) ([]queries.AdminCountUsersByRoleRow, error) {
	return r.queries.AdminCountUsersByRole(ctx)
}

func (r *PostgresRepository) GetUser(ctx context.Context, id pgtype.UUID) (queries.AdminGetUserRow, error) {
	item, err := r.queries.AdminGetUser(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.AdminGetUserRow{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListAuditEvents(ctx context.Context, userID pgtype.UUID) ([]queries.ListAuditEventsForUserRow, error) {
	return r.queries.ListAuditEventsForUser(ctx, userID)
}

// inTx runs fn in a transaction and commits only if it returns nil.
func (r *PostgresRepository) inTx(ctx context.Context, fn func(q *queries.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(r.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func writeAudit(ctx context.Context, q *queries.Queries, audit AuditEntry) error {
	data, err := json.Marshal(audit.Data)
	if err != nil || audit.Data == nil {
		data = []byte("{}")
	}
	return q.CreateAuditEvent(ctx, queries.CreateAuditEventParams{ActorID: audit.ActorID, Action: audit.Action, TargetUserID: audit.Target, Data: data})
}

func (r *PostgresRepository) UpdateUserRole(ctx context.Context, params queries.AdminUpdateUserRoleParams, audit AuditEntry) (queries.AdminUpdateUserRoleRow, error) {
	var item queries.AdminUpdateUserRoleRow
	err := r.inTx(ctx, func(q *queries.Queries) error {
		var err error
		item, err = q.AdminUpdateUserRole(ctx, params)
		if errors.Is(err, pgx.ErrNoRows) {
			// The user exists (checked by the service), so the last-admin guard blocked the update.
			return ErrLastAdmin
		}
		if err != nil {
			return err
		}
		// Force a new login so the next session carries the new role.
		if err := q.RevokeAllRefreshTokensByUser(ctx, params.ID); err != nil {
			return err
		}
		return writeAudit(ctx, q, audit)
	})
	return item, err
}

func (r *PostgresRepository) SuspendUser(ctx context.Context, params queries.AdminSuspendUserParams, audit AuditEntry) error {
	return r.inTx(ctx, func(q *queries.Queries) error {
		if _, err := q.AdminSuspendUser(ctx, params); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLastAdmin
			}
			return err
		}
		if err := q.RevokeAllRefreshTokensByUser(ctx, params.ID); err != nil {
			return err
		}
		return writeAudit(ctx, q, audit)
	})
}

func (r *PostgresRepository) UnsuspendUser(ctx context.Context, id pgtype.UUID, audit AuditEntry) error {
	return r.inTx(ctx, func(q *queries.Queries) error {
		if _, err := q.AdminUnsuspendUser(ctx, id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotSuspended
			}
			return err
		}
		return writeAudit(ctx, q, audit)
	})
}

func (r *PostgresRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.queries.AdminUserExistsByEmail(ctx, email)
}

func (r *PostgresRepository) PendingInviteExists(ctx context.Context, email string) (bool, error) {
	return r.queries.PendingUserInviteExists(ctx, email)
}

func (r *PostgresRepository) CreateInvite(ctx context.Context, params queries.CreateUserInviteParams, audit AuditEntry) (queries.UserInvite, error) {
	var invite queries.UserInvite
	err := r.inTx(ctx, func(q *queries.Queries) error {
		var err error
		if invite, err = q.CreateUserInvite(ctx, params); err != nil {
			return err
		}
		return writeAudit(ctx, q, audit)
	})
	return invite, err
}

func (r *PostgresRepository) ListPendingInvites(ctx context.Context) ([]queries.ListPendingUserInvitesRow, error) {
	return r.queries.ListPendingUserInvites(ctx)
}

func (r *PostgresRepository) GetInviteByID(ctx context.Context, id pgtype.UUID) (queries.UserInvite, error) {
	item, err := r.queries.GetUserInviteByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.UserInvite{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) GetInviteByTokenHash(ctx context.Context, tokenHash string) (queries.UserInvite, error) {
	item, err := r.queries.GetUserInviteByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.UserInvite{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) RotateInvite(ctx context.Context, params queries.RotateUserInviteTokenParams, audit AuditEntry) (queries.UserInvite, error) {
	var invite queries.UserInvite
	err := r.inTx(ctx, func(q *queries.Queries) error {
		var err error
		invite, err = q.RotateUserInviteToken(ctx, params)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteInvalid
		}
		if err != nil {
			return err
		}
		return writeAudit(ctx, q, audit)
	})
	return invite, err
}

func (r *PostgresRepository) RevokeInvite(ctx context.Context, id pgtype.UUID, audit AuditEntry) (queries.UserInvite, error) {
	var invite queries.UserInvite
	err := r.inTx(ctx, func(q *queries.Queries) error {
		var err error
		invite, err = q.RevokeUserInvite(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteInvalid
		}
		if err != nil {
			return err
		}
		return writeAudit(ctx, q, audit)
	})
	return invite, err
}

func (r *PostgresRepository) AcceptInvite(ctx context.Context, inviteID pgtype.UUID, user queries.AdminCreateInvitedUserParams) (pgtype.UUID, error) {
	var userID pgtype.UUID
	err := r.inTx(ctx, func(q *queries.Queries) error {
		invite, err := q.AcceptUserInvite(ctx, inviteID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteInvalid
		}
		if err != nil {
			return err
		}
		if userID, err = q.AdminCreateInvitedUser(ctx, user); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrEmailInUse
			}
			return err
		}
		return writeAudit(ctx, q, AuditEntry{ActorID: invite.InvitedBy, Action: "invite_accepted", Target: userID, Data: map[string]any{"role": invite.Role, "email": invite.Email}})
	})
	return userID, err
}
