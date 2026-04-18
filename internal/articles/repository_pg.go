package articles

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

func (r *PostgresRepository) ListArticles(ctx context.Context, params queries.ListArticlesParams, filters Filters) ([]queries.Article, error) {
	query, args := buildArticleListQuery(filters, false)
	args = append(args, params.Limit, params.Offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]queries.Article, 0)
	for rows.Next() {
		var item queries.Article
		if err := rows.Scan(
			&item.ID,
			&item.Slug,
			&item.Title,
			&item.Excerpt,
			&item.Content,
			&item.CoverImageUrl,
			&item.Type,
			&item.Status,
			&item.SourceName,
			&item.SourceUrl,
			&item.SeoTitle,
			&item.SeoDescription,
			&item.IsFeatured,
			&item.AuthorID,
			&item.PublishedAt,
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

func (r *PostgresRepository) CountArticles(ctx context.Context, filters Filters) (int64, error) {
	query, args := buildArticleListQuery(filters, true)
	row := r.pool.QueryRow(ctx, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PostgresRepository) GetArticleBySlug(ctx context.Context, slug string) (queries.Article, error) {
	item, err := r.queries.GetArticleBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func buildArticleListQuery(filters Filters, count bool) (string, []any) {
	selectClause := `SELECT a.id, a.slug, a.title, a.excerpt, a.content, a.cover_image_url, a.type, a.status, a.source_name, a.source_url, a.seo_title, a.seo_description, a.is_featured, a.author_id, a.published_at, a.created_at, a.updated_at, a.deleted_at FROM articles a`
	if count {
		selectClause = `SELECT COUNT(*) FROM articles a`
	}
	conditions := []string{"a.status = 'published'", "a.deleted_at IS NULL"}
	args := make([]any, 0)
	if filters.Query != "" {
		args = append(args, "%"+filters.Query+"%")
		conditions = append(conditions, fmt.Sprintf("(a.title ILIKE $%d OR COALESCE(a.excerpt, '') ILIKE $%d OR a.content ILIKE $%d)", len(args), len(args), len(args)))
	}
	if filters.Type != "" {
		args = append(args, filters.Type)
		conditions = append(conditions, fmt.Sprintf("a.type = $%d", len(args)))
	}
	if filters.Featured != nil {
		args = append(args, *filters.Featured)
		conditions = append(conditions, fmt.Sprintf("a.is_featured = $%d", len(args)))
	}
	query := selectClause + " WHERE " + strings.Join(conditions, " AND ")
	if count {
		return query, args
	}
	query += fmt.Sprintf(" ORDER BY a.published_at DESC, a.created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	return query, args
}
