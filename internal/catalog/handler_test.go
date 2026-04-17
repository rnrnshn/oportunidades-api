package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

func TestHandlerListUniversities(t *testing.T) {
	universityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listUniversitiesFn: func(context.Context, queries.ListUniversitiesParams, UniversityFilters) ([]queries.University, error) {
			return []queries.University{{
				ID:       pgUUID(universityID),
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
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/catalog/universities", handler.ListUniversities)

	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/universities?page=1&per_page=10", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerGetUniversityBySlugNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{
		getUniversityBySlugFn: func(context.Context, string) (queries.University, error) {
			return queries.University{}, ErrNotFound
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/catalog/universities/:slug", handler.GetUniversityBySlug)

	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/universities/missing", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerListCourses(t *testing.T) {
	courseID := uuid.New()
	universityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listCoursesFn: func(context.Context, queries.ListCoursesParams, CourseFilters) ([]queries.Course, error) {
			return []queries.Course{{
				ID:           pgUUID(courseID),
				UniversityID: pgUUID(universityID),
				Slug:         "informatica",
				Name:         "Informatica",
				Area:         "Tecnologia",
				Level:        "licenciatura",
				Regime:       "presencial",
			}}, nil
		},
		countCoursesFn: func(context.Context, CourseFilters) (int64, error) {
			return 1, nil
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/catalog/courses", handler.ListCourses)

	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/courses", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerListUniversitiesParsesFilters(t *testing.T) {
	called := false
	handler := NewHandler(NewService(&mockRepository{
		listUniversitiesFn: func(_ context.Context, _ queries.ListUniversitiesParams, filters UniversityFilters) ([]queries.University, error) {
			called = true
			if filters.Query != "uem" || filters.Province != "Maputo" || filters.Type != "publica" || filters.Verified == nil || !*filters.Verified {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []queries.University{}, nil
		},
		countUniversitiesFn: func(context.Context, UniversityFilters) (int64, error) {
			return 0, nil
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/catalog/universities", handler.ListUniversities)

	req := httptest.NewRequest(http.MethodGet, "/v1/catalog/universities?q=uem&province=Maputo&type=publica&verified=true", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !called {
		t.Fatal("expected repository call")
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
