package cms

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type PostgresRepository struct{ queries *queries.Queries }

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{queries: queries.New(pool)}
}

func (r *PostgresRepository) CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error) {
	return r.queries.CreateArticle(ctx, params)
}

func (r *PostgresRepository) CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error) {
	return r.queries.CreateOpportunity(ctx, params)
}

func (r *PostgresRepository) ListCMSArticles(ctx context.Context, params queries.ListCMSArticlesParams) ([]queries.Article, error) {
	return r.queries.ListCMSArticles(ctx, params)
}

func (r *PostgresRepository) CountCMSArticles(ctx context.Context) (int64, error) {
	return r.queries.CountCMSArticles(ctx)
}

func (r *PostgresRepository) CreateUniversity(ctx context.Context, params queries.CreateUniversityParams) (queries.University, error) {
	return r.queries.CreateUniversity(ctx, params)
}

func (r *PostgresRepository) ListCMSUniversities(ctx context.Context, params queries.ListCMSUniversitiesParams) ([]queries.University, error) {
	return r.queries.ListCMSUniversities(ctx, params)
}

func (r *PostgresRepository) CountCMSUniversities(ctx context.Context) (int64, error) {
	return r.queries.CountCMSUniversities(ctx)
}

func (r *PostgresRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	item, err := r.queries.GetArticleByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Article{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error) {
	return r.queries.UpdateArticle(ctx, params)
}

func (r *PostgresRepository) GetUniversityByID(ctx context.Context, id pgtype.UUID) (queries.University, error) {
	item, err := r.queries.GetUniversityByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.University{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateUniversity(ctx context.Context, params queries.UpdateUniversityParams) (queries.University, error) {
	return r.queries.UpdateUniversity(ctx, params)
}

func (r *PostgresRepository) CreateCourse(ctx context.Context, params queries.CreateCourseParams) (queries.Course, error) {
	return r.queries.CreateCourse(ctx, params)
}

func (r *PostgresRepository) ListCMSCourses(ctx context.Context, params queries.ListCMSCoursesParams) ([]queries.Course, error) {
	return r.queries.ListCMSCourses(ctx, params)
}

func (r *PostgresRepository) CountCMSCourses(ctx context.Context) (int64, error) {
	return r.queries.CountCMSCourses(ctx)
}

func (r *PostgresRepository) GetCourseByID(ctx context.Context, id pgtype.UUID) (queries.Course, error) {
	item, err := r.queries.GetCourseByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Course{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateCourse(ctx context.Context, params queries.UpdateCourseParams) (queries.Course, error) {
	return r.queries.UpdateCourse(ctx, params)
}

func (r *PostgresRepository) ListCMSOpportunities(ctx context.Context, params queries.ListCMSOpportunitiesParams) ([]queries.Opportunity, error) {
	return r.queries.ListCMSOpportunities(ctx, params)
}

func (r *PostgresRepository) CountCMSOpportunities(ctx context.Context) (int64, error) {
	return r.queries.CountCMSOpportunities(ctx)
}

func (r *PostgresRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	item, err := r.queries.GetOpportunityByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return queries.Opportunity{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error) {
	return r.queries.UpdateOpportunity(ctx, params)
}
