package saved

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	published bool
	saved     []queries.SaveOpportunityParams
	unsaved   []queries.UnsaveOpportunityParams
	rows      []queries.ListSavedOpportunitiesRow
	listArgs  queries.ListSavedOpportunitiesParams
	total     int64
	ids       []pgtype.UUID
}

func (m *mockRepository) IsPublishedOpportunity(context.Context, pgtype.UUID) (bool, error) {
	return m.published, nil
}
func (m *mockRepository) SaveOpportunity(_ context.Context, p queries.SaveOpportunityParams) error {
	m.saved = append(m.saved, p)
	return nil
}
func (m *mockRepository) UnsaveOpportunity(_ context.Context, p queries.UnsaveOpportunityParams) error {
	m.unsaved = append(m.unsaved, p)
	return nil
}
func (m *mockRepository) ListSavedOpportunities(_ context.Context, p queries.ListSavedOpportunitiesParams) ([]queries.ListSavedOpportunitiesRow, error) {
	m.listArgs = p
	return m.rows, nil
}
func (m *mockRepository) CountSavedOpportunities(context.Context, pgtype.UUID) (int64, error) {
	return m.total, nil
}
func (m *mockRepository) ListSavedOpportunityIDs(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
	return m.ids, nil
}

func pgID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

func TestSavePublishedOpportunity(t *testing.T) {
	repo := &mockRepository{published: true}
	userID, oppID := uuid.New(), uuid.New()

	if err := NewService(repo).Save(context.Background(), userID.String(), oppID.String()); err != nil {
		t.Fatalf("save returned error: %v", err)
	}
	if len(repo.saved) != 1 || repo.saved[0].UserID != pgID(userID) || repo.saved[0].OpportunityID != pgID(oppID) {
		t.Fatalf("unexpected save calls: %+v", repo.saved)
	}
}

func TestSaveRejectsUnpublishedOrInvalid(t *testing.T) {
	repo := &mockRepository{published: false}
	service := NewService(repo)

	if err := service.Save(context.Background(), uuid.NewString(), uuid.NewString()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found for unpublished, got %v", err)
	}
	if err := service.Save(context.Background(), uuid.NewString(), "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found for invalid id, got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatal("nothing should have been saved")
	}
}

func TestListMapsOpportunitiesAndClampsPagination(t *testing.T) {
	oppID := uuid.New()
	repo := &mockRepository{
		rows:  []queries.ListSavedOpportunitiesRow{{Opportunity: queries.Opportunity{ID: pgID(oppID), Slug: "bolsa", Title: "Bolsa", IsActive: true, Verified: true}}},
		total: 1,
	}

	result, err := NewService(repo).List(context.Background(), uuid.NewString(), 0, 500)
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if len(result.Data) != 1 || result.Data[0].ID != oppID.String() || result.Data[0].Slug != "bolsa" {
		t.Fatalf("unexpected data: %+v", result.Data)
	}
	if repo.listArgs.Limit != maxPerPage || repo.listArgs.Offset != 0 || result.Meta.Page != 1 || result.Meta.TotalPages != 1 {
		t.Fatalf("unexpected pagination: args=%+v meta=%+v", repo.listArgs, result.Meta)
	}
}

func TestIDs(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	result, err := NewService(&mockRepository{ids: []pgtype.UUID{pgID(a), pgID(b)}}).IDs(context.Background(), uuid.NewString())
	if err != nil {
		t.Fatalf("ids returned error: %v", err)
	}
	if len(result.Data) != 2 || result.Data[0] != a.String() || result.Data[1] != b.String() {
		t.Fatalf("unexpected ids: %+v", result.Data)
	}
}

func TestHandlerSaveStatusCodes(t *testing.T) {
	cases := map[string]struct {
		published bool
		want      int
	}{
		"published":   {true, http.StatusNoContent},
		"unpublished": {false, http.StatusNotFound},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(NewService(&mockRepository{published: tc.published}))
			app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
			app.Put("/saved/:opportunityId", func(c *fiber.Ctx) error {
				c.Locals("auth_user", appauth.AuthenticatedUser{ID: uuid.NewString(), Role: "user"})
				return handler.Save(c)
			})
			res, err := app.Test(httptest.NewRequest(http.MethodPut, "/saved/"+uuid.NewString(), nil))
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if res.StatusCode != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, res.StatusCode)
			}
		})
	}
}
