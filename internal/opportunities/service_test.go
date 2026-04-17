package opportunities

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	listOpportunitiesFn    func(context.Context, queries.ListOpportunitiesParams, Filters) ([]queries.Opportunity, error)
	countOpportunitiesFn   func(context.Context, Filters) (int64, error)
	getOpportunityBySlugFn func(context.Context, string) (queries.Opportunity, error)
}

func (m *mockRepository) ListOpportunities(ctx context.Context, params queries.ListOpportunitiesParams, filters Filters) ([]queries.Opportunity, error) {
	return m.listOpportunitiesFn(ctx, params, filters)
}

func (m *mockRepository) CountOpportunities(ctx context.Context, filters Filters) (int64, error) {
	return m.countOpportunitiesFn(ctx, filters)
}

func (m *mockRepository) GetOpportunityBySlug(ctx context.Context, slug string) (queries.Opportunity, error) {
	return m.getOpportunityBySlugFn(ctx, slug)
}

func TestListOpportunities(t *testing.T) {
	opportunityID := uuid.New()
	service := NewService(&mockRepository{
		listOpportunitiesFn: func(context.Context, queries.ListOpportunitiesParams, Filters) ([]queries.Opportunity, error) {
			return []queries.Opportunity{{
				ID:          uuidToPg(opportunityID),
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
	})

	result, err := service.ListOpportunities(context.Background(), PaginationParams{}, Filters{})
	if err != nil {
		t.Fatalf("list opportunities returned error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 opportunity, got %d", len(result.Data))
	}

	if result.Meta.Total != 1 || result.Meta.Page != 1 || result.Meta.PerPage != 20 {
		t.Fatalf("unexpected meta: %+v", result.Meta)
	}
}

func TestGetOpportunityBySlugNotFound(t *testing.T) {
	service := NewService(&mockRepository{
		getOpportunityBySlugFn: func(context.Context, string) (queries.Opportunity, error) {
			return queries.Opportunity{}, ErrNotFound
		},
	})

	_, err := service.GetOpportunityBySlug(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetOpportunityBySlug(t *testing.T) {
	opportunityID := uuid.New()
	service := NewService(&mockRepository{
		getOpportunityBySlugFn: func(context.Context, string) (queries.Opportunity, error) {
			return queries.Opportunity{
				ID:          uuidToPg(opportunityID),
				Slug:        "estagio-test",
				Title:       "Estagio Test",
				Type:        "estagio",
				EntityName:  "Empresa Teste",
				Description: "Descricao",
				Country:     "Mozambique",
				Deadline:    pgtype.Timestamptz{Time: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC), Valid: true},
			}, nil
		},
	})

	result, err := service.GetOpportunityBySlug(context.Background(), "estagio-test")
	if err != nil {
		t.Fatalf("get opportunity returned error: %v", err)
	}

	if result.Data.Deadline == "" {
		t.Fatal("expected deadline to be mapped")
	}
}

func uuidToPg(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}
