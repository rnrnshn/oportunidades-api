package saved

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rnrnshn/oportunidades-api/internal/opportunities"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var ErrNotFound = errors.New("saved: opportunity not found")

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

type Service struct {
	repo Repository
}

type ListResult struct {
	Data []opportunities.OpportunityItem `json:"data"`
	Meta opportunities.PaginationMeta    `json:"meta"`
}

type IDsResult struct {
	Data []string `json:"data"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(ctx context.Context, userID string, opportunityID string) error {
	user, opportunity, err := parseIDs(userID, opportunityID)
	if err != nil {
		return err
	}
	published, err := s.repo.IsPublishedOpportunity(ctx, opportunity)
	if err != nil {
		return fmt.Errorf("saved: check opportunity: %w", err)
	}
	if !published {
		return ErrNotFound
	}
	if err := s.repo.SaveOpportunity(ctx, queries.SaveOpportunityParams{UserID: user, OpportunityID: opportunity}); err != nil {
		return fmt.Errorf("saved: save opportunity: %w", err)
	}
	return nil
}

func (s *Service) Unsave(ctx context.Context, userID string, opportunityID string) error {
	user, opportunity, err := parseIDs(userID, opportunityID)
	if err != nil {
		return err
	}
	if err := s.repo.UnsaveOpportunity(ctx, queries.UnsaveOpportunityParams{UserID: user, OpportunityID: opportunity}); err != nil {
		return fmt.Errorf("saved: unsave opportunity: %w", err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, userID string, page int, perPage int) (*ListResult, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	rows, err := s.repo.ListSavedOpportunities(ctx, queries.ListSavedOpportunitiesParams{
		UserID: user,
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("saved: list: %w", err)
	}
	total, err := s.repo.CountSavedOpportunities(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("saved: count: %w", err)
	}

	data := make([]opportunities.OpportunityItem, 0, len(rows))
	for _, row := range rows {
		item, err := opportunities.MapOpportunity(row.Opportunity)
		if err != nil {
			return nil, err
		}
		data = append(data, item)
	}

	return &ListResult{
		Data: data,
		Meta: opportunities.PaginationMeta{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: int(math.Ceil(float64(total) / float64(perPage))),
		},
	}, nil
}

func (s *Service) IDs(ctx context.Context, userID string) (*IDsResult, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	ids, err := s.repo.ListSavedOpportunityIDs(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("saved: list ids: %w", err)
	}
	result := &IDsResult{Data: make([]string, 0, len(ids))}
	for _, id := range ids {
		result.Data = append(result.Data, uuid.UUID(id.Bytes).String())
	}
	return result, nil
}

func parseIDs(userID string, opportunityID string) (pgtype.UUID, pgtype.UUID, error) {
	user, err := parseID(userID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, err
	}
	opportunity, err := parseID(opportunityID)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, ErrNotFound
	}
	return user, opportunity, nil
}

func parseID(value string) (pgtype.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("saved: invalid id: %w", err)
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}
