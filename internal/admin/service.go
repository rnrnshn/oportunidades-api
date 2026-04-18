package admin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type Service struct{ repo Repository }

type ArticleResult struct {
	Data ArticleItem `json:"data"`
}
type OpportunityResult struct {
	Data OpportunityItem `json:"data"`
}

type ArticleItem struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
}

type OpportunityItem struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Verified bool   `json:"verified"`
	IsActive bool   `json:"is_active"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) PublishArticle(ctx context.Context, articleID string) (*ArticleResult, error) {
	id, err := parseID(articleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetArticleByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.PublishArticle(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := mapArticle(item)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

func (s *Service) VerifyOpportunity(ctx context.Context, opportunityID string) (*OpportunityResult, error) {
	id, err := parseID(opportunityID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetOpportunityByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.VerifyOpportunity(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := mapOpportunity(item)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func parseID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}, nil
}

func mapArticle(article queries.Article) (ArticleItem, error) {
	id, err := uuidFromPg(article.ID)
	if err != nil {
		return ArticleItem{}, fmt.Errorf("admin: article id: %w", err)
	}
	return ArticleItem{ID: id.String(), Slug: article.Slug, Title: article.Title, Status: article.Status, PublishedAt: timestamptzValue(article.PublishedAt)}, nil
}

func mapOpportunity(opportunity queries.Opportunity) (OpportunityItem, error) {
	id, err := uuidFromPg(opportunity.ID)
	if err != nil {
		return OpportunityItem{}, fmt.Errorf("admin: opportunity id: %w", err)
	}
	return OpportunityItem{ID: id.String(), Slug: opportunity.Slug, Title: opportunity.Title, Verified: opportunity.Verified, IsActive: opportunity.IsActive}, nil
}

func uuidFromPg(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(value.Bytes), nil
}
func timestamptzValue(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
