package opportunities

import (
	"context"
	"errors"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("opportunities: not found")

type Repository interface {
	ListOpportunities(ctx context.Context, params queries.ListOpportunitiesParams, filters Filters) ([]queries.Opportunity, error)
	CountOpportunities(ctx context.Context, filters Filters) (int64, error)
	GetOpportunityBySlug(ctx context.Context, slug string) (queries.Opportunity, error)
}
