package catalog

import (
	"context"
	"errors"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("catalog: not found")

type Repository interface {
	ListUniversities(ctx context.Context, params queries.ListUniversitiesParams, filters UniversityFilters) ([]queries.University, error)
	CountUniversities(ctx context.Context, filters UniversityFilters) (int64, error)
	GetUniversityBySlug(ctx context.Context, slug string) (queries.University, error)
	ListCourses(ctx context.Context, params queries.ListCoursesParams, filters CourseFilters) ([]queries.Course, error)
	CountCourses(ctx context.Context, filters CourseFilters) (int64, error)
	GetCourseBySlug(ctx context.Context, slug string) (queries.Course, error)
}
