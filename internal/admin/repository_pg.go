package admin

import (
	"context"
	"errors"

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
				(CURRENT_DATE - ($1 || ' days')::interval)::date,
				CURRENT_DATE,
				'1 day'::interval
			)::date AS day
		),
		art AS (
			SELECT created_at::date AS day, COUNT(*) AS cnt
			FROM articles WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE - ($1 || ' days')::interval
			GROUP BY 1
		),
		opp AS (
			SELECT created_at::date AS day, COUNT(*) AS cnt
			FROM opportunities WHERE deleted_at IS NULL AND created_at >= CURRENT_DATE - ($1 || ' days')::interval
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
