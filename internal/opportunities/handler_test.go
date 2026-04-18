package opportunities

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

func TestHandlerListOpportunities(t *testing.T) {
	opportunityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		listOpportunitiesFn: func(context.Context, queries.ListOpportunitiesParams, Filters) ([]queries.Opportunity, error) {
			return []queries.Opportunity{{
				ID:          pgUUID(opportunityID),
				Slug:        "bolsa-test",
				Title:       "Bolsa Test",
				Type:        "bolsa",
				EntityName:  "Universidade Teste",
				Description: "Descricao",
				Country:     "Mozambique",
				IsActive:    true,
				Verified:    true,
			}}, nil
		},
		countOpportunitiesFn: func(context.Context, Filters) (int64, error) {
			return 1, nil
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/opportunities", handler.ListOpportunities)

	req := httptest.NewRequest(http.MethodGet, "/v1/opportunities", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerGetOpportunityBySlugNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{
		getOpportunityBySlugFn: func(context.Context, string) (queries.Opportunity, error) {
			return queries.Opportunity{}, ErrNotFound
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/opportunities/:slug", handler.GetOpportunityBySlug)

	req := httptest.NewRequest(http.MethodGet, "/v1/opportunities/missing", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerListOpportunitiesParsesFilters(t *testing.T) {
	called := false
	handler := NewHandler(NewService(&mockRepository{
		listOpportunitiesFn: func(_ context.Context, _ queries.ListOpportunitiesParams, filters Filters) ([]queries.Opportunity, error) {
			called = true
			if filters.Query != "bolsa" || filters.Country != "Mozambique" || filters.Type != "bolsa" || filters.Active == nil || !*filters.Active || filters.Verified == nil || !*filters.Verified {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []queries.Opportunity{}, nil
		},
		countOpportunitiesFn: func(context.Context, Filters) (int64, error) {
			return 0, nil
		},
	}))

	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/opportunities", handler.ListOpportunities)

	req := httptest.NewRequest(http.MethodGet, "/v1/opportunities?q=bolsa&type=bolsa&country=Mozambique&active=true&verified=true", nil)
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

func TestHandlerListOpportunitiesValidatesQuery(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/opportunities", handler.ListOpportunities)
	req := httptest.NewRequest(http.MethodGet, "/v1/opportunities?type=bad&active=maybe&verified=nah", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
