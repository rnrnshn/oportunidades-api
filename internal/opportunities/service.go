package opportunities

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

type Service struct {
	repo Repository
}

type PaginationParams struct {
	Page    int
	PerPage int
}

type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

type Filters struct {
	Query    string
	Type     string
	Area     string
	Country  string
	Active   *bool
	Verified *bool
}

type OpportunitiesResult struct {
	Data []OpportunityItem `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type OpportunityDetailResult struct {
	Data OpportunityItem `json:"data"`
}

type OpportunityItem struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Type         string `json:"type"`
	EntityName   string `json:"entity_name"`
	Description  string `json:"description"`
	Requirements string `json:"requirements,omitempty"`
	Deadline     string `json:"deadline,omitempty"`
	ApplyURL     string `json:"apply_url,omitempty"`
	Country      string `json:"country"`
	Language     string `json:"language,omitempty"`
	Area         string `json:"area,omitempty"`
	IsActive     bool   `json:"is_active"`
	Verified     bool   `json:"verified"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListOpportunities(ctx context.Context, params PaginationParams, filters Filters) (*OpportunitiesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListOpportunities(ctx, queries.ListOpportunitiesParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	}, filters)
	if err != nil {
		return nil, fmt.Errorf("opportunities: list opportunities: %w", err)
	}

	total, err := s.repo.CountOpportunities(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("opportunities: count opportunities: %w", err)
	}

	data := make([]OpportunityItem, 0, len(items))
	for _, item := range items {
		mappedItem, err := mapOpportunity(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mappedItem)
	}

	return &OpportunitiesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetOpportunityBySlug(ctx context.Context, slug string) (*OpportunityDetailResult, error) {
	item, err := s.repo.GetOpportunityBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	mappedItem, err := mapOpportunity(item)
	if err != nil {
		return nil, err
	}

	return &OpportunityDetailResult{Data: mappedItem}, nil
}

func normalizePagination(params PaginationParams) (int, int) {
	page := params.Page
	if page < 1 {
		page = defaultPage
	}

	perPage := params.PerPage
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage
}

func buildMeta(total int64, page int, perPage int) PaginationMeta {
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	return PaginationMeta{Total: total, Page: page, PerPage: perPage, TotalPages: totalPages}
}

func mapOpportunity(item queries.Opportunity) (OpportunityItem, error) {
	id, err := uuidFromPg(item.ID)
	if err != nil {
		return OpportunityItem{}, fmt.Errorf("opportunities: id: %w", err)
	}

	return OpportunityItem{
		ID:           id.String(),
		Slug:         item.Slug,
		Title:        item.Title,
		Type:         item.Type,
		EntityName:   item.EntityName,
		Description:  item.Description,
		Requirements: textValue(item.Requirements),
		Deadline:     timestamptzValue(item.Deadline),
		ApplyURL:     textValue(item.ApplyUrl),
		Country:      item.Country,
		Language:     textValue(item.Language),
		Area:         textValue(item.Area),
		IsActive:     item.IsActive,
		Verified:     item.Verified,
	}, nil
}

func uuidFromPg(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}

	return uuid.UUID(value.Bytes), nil
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func timestamptzValue(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}

	return value.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
}
