package saved

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Repository interface {
	IsPublishedOpportunity(ctx context.Context, id pgtype.UUID) (bool, error)
	SaveOpportunity(ctx context.Context, params queries.SaveOpportunityParams) error
	UnsaveOpportunity(ctx context.Context, params queries.UnsaveOpportunityParams) error
	ListSavedOpportunities(ctx context.Context, params queries.ListSavedOpportunitiesParams) ([]queries.ListSavedOpportunitiesRow, error)
	CountSavedOpportunities(ctx context.Context, userID pgtype.UUID) (int64, error)
	ListSavedOpportunityIDs(ctx context.Context, userID pgtype.UUID) ([]pgtype.UUID, error)
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return queries.New(pool)
}
