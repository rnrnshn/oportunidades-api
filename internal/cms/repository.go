package cms

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Repository interface {
	CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error)
	CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error)
	GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error)
	UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error)
	GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error)
	UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error)
}
