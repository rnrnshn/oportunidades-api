package cms

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("cms: not found")
var ErrForbidden = errors.New("cms: forbidden")

type Repository interface {
	CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error)
	CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error)
	ListCMSArticles(ctx context.Context, params queries.ListCMSArticlesParams, actor Actor, filters ArticleListFilters) ([]queries.Article, error)
	CountCMSArticles(ctx context.Context, actor Actor, filters ArticleListFilters) (int64, error)
	GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error)
	CreateUniversity(ctx context.Context, params queries.CreateUniversityParams) (queries.University, error)
	ListCMSUniversities(ctx context.Context, params queries.ListCMSUniversitiesParams, actor Actor, filters UniversityListFilters) ([]queries.University, error)
	CountCMSUniversities(ctx context.Context, actor Actor, filters UniversityListFilters) (int64, error)
	GetUniversityByID(ctx context.Context, id pgtype.UUID) (queries.University, error)
	UpdateUniversity(ctx context.Context, params queries.UpdateUniversityParams) (queries.University, error)
	CreateCourse(ctx context.Context, params queries.CreateCourseParams) (queries.Course, error)
	ListCMSCourses(ctx context.Context, params queries.ListCMSCoursesParams, actor Actor, filters CourseListFilters) ([]queries.Course, error)
	CountCMSCourses(ctx context.Context, actor Actor, filters CourseListFilters) (int64, error)
	GetCourseByID(ctx context.Context, id pgtype.UUID) (queries.Course, error)
	UpdateCourse(ctx context.Context, params queries.UpdateCourseParams) (queries.Course, error)
	ListCMSOpportunities(ctx context.Context, params queries.ListCMSOpportunitiesParams, actor Actor, filters OpportunityListFilters) ([]queries.Opportunity, error)
	CountCMSOpportunities(ctx context.Context, actor Actor, filters OpportunityListFilters) (int64, error)
	GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error)
}
