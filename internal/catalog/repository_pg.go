package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool, queries: queries.New(pool)}
}

func (r *PostgresRepository) ListUniversities(ctx context.Context, params queries.ListUniversitiesParams, filters UniversityFilters) ([]queries.University, error) {
	query, args := buildUniversityListQuery(filters, false)
	args = append(args, params.Limit, params.Offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]queries.University, 0)
	for rows.Next() {
		var item queries.University
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Name,
			&item.Type,
			&item.Province,
			&item.Description,
			&item.LogoUrl,
			&item.Website,
			&item.Email,
			&item.Phone,
			&item.Verified,
			&item.VerifiedAt,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresRepository) CountUniversities(ctx context.Context, filters UniversityFilters) (int64, error) {
	query, args := buildUniversityListQuery(filters, true)
	row := r.pool.QueryRow(ctx, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostgresRepository) GetUniversityBySlug(ctx context.Context, slug string) (queries.University, error) {
	item, err := r.queries.GetUniversityBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.University{}, ErrNotFound
	}

	return item, err
}

func (r *PostgresRepository) ListCourses(ctx context.Context, params queries.ListCoursesParams, filters CourseFilters) ([]queries.Course, error) {
	query, args := buildCourseListQuery(filters, false)
	args = append(args, params.Limit, params.Offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]queries.Course, 0)
	for rows.Next() {
		var item queries.Course
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.UniversityID,
			&item.Name,
			&item.Area,
			&item.Level,
			&item.Regime,
			&item.DurationYears,
			&item.AnnualFee,
			&item.EntryRequirements,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *PostgresRepository) CountCourses(ctx context.Context, filters CourseFilters) (int64, error) {
	query, args := buildCourseListQuery(filters, true)
	row := r.pool.QueryRow(ctx, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostgresRepository) GetCourseBySlug(ctx context.Context, slug string) (queries.Course, error) {
	item, err := r.queries.GetCourseBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Course{}, ErrNotFound
	}

	return item, err
}

func buildUniversityListQuery(filters UniversityFilters, count bool) (string, []any) {
	selectClause := `SELECT u.id, u.slug, u.name, u.type, u.province, u.description, u.logo_url, u.website, u.email, u.phone, u.verified, u.verified_at, u.created_by, u.created_at, u.updated_at, u.deleted_at FROM universities u`
	if count {
		selectClause = `SELECT COUNT(*) FROM universities u`
	}

	conditions := []string{"u.deleted_at IS NULL"}
	args := make([]any, 0)

	if filters.Query != "" {
		args = append(args, "%"+filters.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(u.name ILIKE $%d OR COALESCE(u.description, '') ILIKE $%d)", len(args), len(args)))
	}
	if filters.Province != "" {
		args = append(args, filters.Province)
		conditions = append(conditions, fmt.Sprintf("u.province = $%d", len(args)))
	}
	if filters.Type != "" {
		args = append(args, filters.Type)
		conditions = append(conditions, fmt.Sprintf("u.type = $%d", len(args)))
	}
	if filters.Verified != nil {
		args = append(args, *filters.Verified)
		conditions = append(conditions, fmt.Sprintf("u.verified = $%d", len(args)))
	}

	query := selectClause + " WHERE " + strings.Join(conditions, " AND ")
	if count {
		return query, args
	}

	query += fmt.Sprintf(" ORDER BY u.name ASC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	return query, args
}

func buildCourseListQuery(filters CourseFilters, count bool) (string, []any) {
	selectClause := `SELECT c.id, c.slug, c.university_id, c.name, c.area, c.level, c.regime, c.duration_years, c.annual_fee, c.entry_requirements, c.created_at, c.updated_at, c.deleted_at FROM courses c JOIN universities u ON u.id = c.university_id`
	if count {
		selectClause = `SELECT COUNT(*) FROM courses c JOIN universities u ON u.id = c.university_id`
	}

	conditions := []string{"c.deleted_at IS NULL", "u.deleted_at IS NULL"}
	args := make([]any, 0)

	if filters.Query != "" {
		args = append(args, "%"+filters.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(c.name ILIKE $%d OR c.area ILIKE $%d OR u.name ILIKE $%d)", len(args), len(args), len(args)))
	}
	if filters.Area != "" {
		args = append(args, filters.Area)
		conditions = append(conditions, fmt.Sprintf("c.area = $%d", len(args)))
	}
	if filters.Level != "" {
		args = append(args, filters.Level)
		conditions = append(conditions, fmt.Sprintf("c.level = $%d", len(args)))
	}
	if filters.Regime != "" {
		args = append(args, filters.Regime)
		conditions = append(conditions, fmt.Sprintf("c.regime = $%d", len(args)))
	}
	if filters.Province != "" {
		args = append(args, filters.Province)
		conditions = append(conditions, fmt.Sprintf("u.province = $%d", len(args)))
	}
	if filters.UniversityID != "" {
		args = append(args, filters.UniversityID)
		conditions = append(conditions, fmt.Sprintf("c.university_id::text = $%d", len(args)))
	}

	query := selectClause + " WHERE " + strings.Join(conditions, " AND ")
	if count {
		return query, args
	}

	query += fmt.Sprintf(" ORDER BY c.name ASC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	return query, args
}
