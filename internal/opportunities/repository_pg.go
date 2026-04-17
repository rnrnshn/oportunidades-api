package opportunities

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

func (r *PostgresRepository) ListOpportunities(ctx context.Context, params queries.ListOpportunitiesParams, filters Filters) ([]queries.Opportunity, error) {
	query, args := buildOpportunityListQuery(filters, false)
	args = append(args, params.Limit, params.Offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]queries.Opportunity, 0)
	for rows.Next() {
		var item queries.Opportunity
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Type,
			&item.EntityName,
			&item.Description,
			&item.Requirements,
			&item.Deadline,
			&item.ApplyUrl,
			&item.Country,
			&item.Language,
			&item.Area,
			&item.IsActive,
			&item.PublishedBy,
			&item.Verified,
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

func (r *PostgresRepository) CountOpportunities(ctx context.Context, filters Filters) (int64, error) {
	query, args := buildOpportunityListQuery(filters, true)
	row := r.pool.QueryRow(ctx, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (r *PostgresRepository) GetOpportunityBySlug(ctx context.Context, slug string) (queries.Opportunity, error) {
	item, err := r.queries.GetOpportunityBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}

	return item, err
}

func buildOpportunityListQuery(filters Filters, count bool) (string, []any) {
	selectClause := `SELECT o.id, o.slug, o.title, o.type, o.entity_name, o.description, o.requirements, o.deadline, o.apply_url, o.country, o.language, o.area, o.is_active, o.published_by, o.verified, o.created_at, o.updated_at, o.deleted_at FROM opportunities o`
	if count {
		selectClause = `SELECT COUNT(*) FROM opportunities o`
	}

	conditions := []string{"o.deleted_at IS NULL"}
	args := make([]any, 0)

	if filters.Query != "" {
		args = append(args, "%"+filters.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(o.title ILIKE $%d OR o.description ILIKE $%d OR o.entity_name ILIKE $%d)", len(args), len(args), len(args)))
	}
	if filters.Type != "" {
		args = append(args, filters.Type)
		conditions = append(conditions, fmt.Sprintf("o.type = $%d", len(args)))
	}
	if filters.Area != "" {
		args = append(args, filters.Area)
		conditions = append(conditions, fmt.Sprintf("o.area = $%d", len(args)))
	}
	if filters.Country != "" {
		args = append(args, filters.Country)
		conditions = append(conditions, fmt.Sprintf("o.country = $%d", len(args)))
	}
	if filters.Active != nil {
		args = append(args, *filters.Active)
		conditions = append(conditions, fmt.Sprintf("o.is_active = $%d", len(args)))
	}
	if filters.Verified != nil {
		args = append(args, *filters.Verified)
		conditions = append(conditions, fmt.Sprintf("o.verified = $%d", len(args)))
	}

	query := selectClause + " WHERE " + strings.Join(conditions, " AND ")
	if count {
		return query, args
	}

	query += fmt.Sprintf(" ORDER BY o.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	return query, args
}
