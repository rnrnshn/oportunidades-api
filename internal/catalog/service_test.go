package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	listUniversitiesFn    func(context.Context, queries.ListUniversitiesParams, UniversityFilters) ([]queries.University, error)
	countUniversitiesFn   func(context.Context, UniversityFilters) (int64, error)
	getUniversityBySlugFn func(context.Context, string) (queries.University, error)
	listCoursesFn         func(context.Context, queries.ListCoursesParams, CourseFilters) ([]queries.Course, error)
	countCoursesFn        func(context.Context, CourseFilters) (int64, error)
	getCourseBySlugFn     func(context.Context, string) (queries.Course, error)
}

func (m *mockRepository) ListUniversities(ctx context.Context, params queries.ListUniversitiesParams, filters UniversityFilters) ([]queries.University, error) {
	return m.listUniversitiesFn(ctx, params, filters)
}

func (m *mockRepository) CountUniversities(ctx context.Context, filters UniversityFilters) (int64, error) {
	return m.countUniversitiesFn(ctx, filters)
}

func (m *mockRepository) GetUniversityBySlug(ctx context.Context, slug string) (queries.University, error) {
	return m.getUniversityBySlugFn(ctx, slug)
}

func (m *mockRepository) ListCourses(ctx context.Context, params queries.ListCoursesParams, filters CourseFilters) ([]queries.Course, error) {
	return m.listCoursesFn(ctx, params, filters)
}

func (m *mockRepository) CountCourses(ctx context.Context, filters CourseFilters) (int64, error) {
	return m.countCoursesFn(ctx, filters)
}

func (m *mockRepository) GetCourseBySlug(ctx context.Context, slug string) (queries.Course, error) {
	return m.getCourseBySlugFn(ctx, slug)
}

func TestListUniversities(t *testing.T) {
	universityID := uuid.New()
	service := NewService(&mockRepository{
		listUniversitiesFn: func(context.Context, queries.ListUniversitiesParams, UniversityFilters) ([]queries.University, error) {
			return []queries.University{{
				ID:       uuidToPg(universityID),
				Slug:     "uem",
				Name:     "UEM",
				Type:     "publica",
				Province: "Maputo",
				Verified: true,
			}}, nil
		},
		countUniversitiesFn: func(context.Context, UniversityFilters) (int64, error) {
			return 1, nil
		},
	})

	result, err := service.ListUniversities(context.Background(), PaginationParams{}, UniversityFilters{})
	if err != nil {
		t.Fatalf("list universities returned error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 university, got %d", len(result.Data))
	}

	if result.Meta.Total != 1 || result.Meta.Page != 1 || result.Meta.PerPage != 20 {
		t.Fatalf("unexpected meta: %+v", result.Meta)
	}
}

func TestGetUniversityBySlugNotFound(t *testing.T) {
	service := NewService(&mockRepository{
		getUniversityBySlugFn: func(context.Context, string) (queries.University, error) {
			return queries.University{}, ErrNotFound
		},
	})

	_, err := service.GetUniversityBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestListCourses(t *testing.T) {
	courseID := uuid.New()
	universityID := uuid.New()
	service := NewService(&mockRepository{
		listCoursesFn: func(context.Context, queries.ListCoursesParams, CourseFilters) ([]queries.Course, error) {
			return []queries.Course{{
				ID:           uuidToPg(courseID),
				UniversityID: uuidToPg(universityID),
				Slug:         "engenharia-informatica",
				Name:         "Engenharia Informatica",
				Area:         "Tecnologia",
				Level:        "licenciatura",
				Regime:       "presencial",
			}}, nil
		},
		countCoursesFn: func(context.Context, CourseFilters) (int64, error) {
			return 1, nil
		},
	})

	result, err := service.ListCourses(context.Background(), PaginationParams{Page: 2, PerPage: 10}, CourseFilters{})
	if err != nil {
		t.Fatalf("list courses returned error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 course, got %d", len(result.Data))
	}

	if result.Meta.Page != 2 || result.Meta.PerPage != 10 {
		t.Fatalf("unexpected meta: %+v", result.Meta)
	}
}

func uuidToPg(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
