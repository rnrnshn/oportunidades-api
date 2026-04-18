package cms

import (
	"context"

	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Repository interface {
	CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error)
	CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error)
}
