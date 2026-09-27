package cms

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

func TestUpdateOpportunityClearsOptionalFields(t *testing.T) {
	id := uuid.New()
	var got queries.UpdateOpportunityParams
	service := NewService(&mockRepository{
		getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{
				ID:          pgtype.UUID{Bytes: id, Valid: true},
				Title:       "Bolsa",
				Type:        "bolsa",
				EntityName:  "Fundação",
				Description: "Descrição",
				Country:     "Moçambique",
				Language:    pgtype.Text{String: "Inglês", Valid: true},
				AmountMax:   numericToPg("1000"),
				Deadline:    pgtype.Timestamptz{Valid: true},
				PublishedBy: pgtype.UUID{Bytes: uuid.New(), Valid: true},
			}, nil
		},
		updateOpportunityFn: func(_ context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error) {
			got = params
			return queries.Opportunity{ID: params.ID, Title: params.Title, Type: params.Type, Country: params.Country, PublishedBy: pgtype.UUID{Bytes: uuid.New(), Valid: true}}, nil
		},
	})

	_, err := service.UpdateOpportunity(context.Background(), Actor{UserID: uuid.NewString(), Role: "admin"}, CreateOpportunityInput{
		ID:          id.String(),
		Title:       "Bolsa",
		Type:        "bolsa",
		EntityName:  "Fundação",
		Description: "Descrição",
		Country:     "Moçambique",
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if got.Language.Valid || got.AmountMax.Valid || got.Deadline.Valid {
		t.Fatalf("expected cleared language, amount and deadline, got %+v %+v %+v", got.Language, got.AmountMax, got.Deadline)
	}
}
