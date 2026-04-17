package mentorship

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

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

type MentorsResult struct {
	Data []MentorItem   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type MentorDetailResult struct {
	Data MentorItem `json:"data"`
}

type SessionRequestInput struct {
	MentorID    string
	RequesterID string
	Message     string
	ScheduledAt string
}

type SessionRequestResult struct {
	Data SessionRequestItem `json:"data"`
}

type MentorItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Headline     string `json:"headline"`
	Bio          string `json:"bio"`
	Expertise    string `json:"expertise"`
	Availability string `json:"availability,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
}

type SessionRequestItem struct {
	ID          string `json:"id"`
	MentorID    string `json:"mentor_id"`
	RequesterID string `json:"requester_id"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	ScheduledAt string `json:"scheduled_at,omitempty"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListMentors(ctx context.Context, params PaginationParams) (*MentorsResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListMentors(ctx, queries.ListMentorsParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	})
	if err != nil {
		return nil, fmt.Errorf("mentorship: list mentors: %w", err)
	}

	total, err := s.repo.CountMentors(ctx)
	if err != nil {
		return nil, fmt.Errorf("mentorship: count mentors: %w", err)
	}

	data := make([]MentorItem, 0, len(items))
	for _, item := range items {
		mappedItem, err := mapMentorListItem(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mappedItem)
	}

	return &MentorsResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetMentorByID(ctx context.Context, mentorID string) (*MentorDetailResult, error) {
	mentorUUID, err := uuid.Parse(strings.TrimSpace(mentorID))
	if err != nil {
		return nil, ErrNotFound
	}

	item, err := s.repo.GetMentorByID(ctx, uuidToPg(mentorUUID))
	if err != nil {
		return nil, err
	}

	mappedItem, err := mapMentorDetailItem(item)
	if err != nil {
		return nil, err
	}

	return &MentorDetailResult{Data: mappedItem}, nil
}

func (s *Service) CreateSessionRequest(ctx context.Context, input SessionRequestInput) (*SessionRequestResult, error) {
	mentorUUID, err := uuid.Parse(strings.TrimSpace(input.MentorID))
	if err != nil {
		return nil, fmt.Errorf("mentorship: invalid mentor id: %w", err)
	}
	requesterUUID, err := uuid.Parse(strings.TrimSpace(input.RequesterID))
	if err != nil {
		return nil, fmt.Errorf("mentorship: invalid requester id: %w", err)
	}
	if mentorUUID == requesterUUID {
		return nil, fmt.Errorf("mentorship: requester cannot book own mentor profile")
	}

	if _, err := s.repo.GetMentorByID(ctx, uuidToPg(mentorUUID)); err != nil {
		return nil, err
	}

	var scheduledAt pgtype.Timestamptz
	if strings.TrimSpace(input.ScheduledAt) != "" {
		parsedTime, err := time.Parse(time.RFC3339, input.ScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("mentorship: invalid scheduled_at: %w", err)
		}
		scheduledAt = pgtype.Timestamptz{Time: parsedTime, Valid: true}
	}

	session, err := s.repo.CreateMentorshipSession(ctx, queries.CreateMentorshipSessionParams{
		MentorID:    uuidToPg(mentorUUID),
		RequesterID: uuidToPg(requesterUUID),
		Message:     strings.TrimSpace(input.Message),
		Status:      "pending",
		ScheduledAt: scheduledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("mentorship: create session request: %w", err)
	}

	mappedSession, err := mapSession(session)
	if err != nil {
		return nil, err
	}

	return &SessionRequestResult{Data: mappedSession}, nil
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

func mapMentorListItem(item queries.ListMentorsRow) (MentorItem, error) {
	id, err := uuidFromPg(item.UserID)
	if err != nil {
		return MentorItem{}, fmt.Errorf("mentorship: mentor user id: %w", err)
	}
	return MentorItem{
		ID:           id.String(),
		Name:         item.Name,
		Email:        item.Email,
		Headline:     item.Headline,
		Bio:          item.Bio,
		Expertise:    item.Expertise,
		Availability: textValue(item.Availability),
		AvatarURL:    textValue(item.AvatarUrl),
	}, nil
}

func mapMentorDetailItem(item queries.GetMentorByIDRow) (MentorItem, error) {
	id, err := uuidFromPg(item.UserID)
	if err != nil {
		return MentorItem{}, fmt.Errorf("mentorship: mentor detail user id: %w", err)
	}
	return MentorItem{
		ID:           id.String(),
		Name:         item.Name,
		Email:        item.Email,
		Headline:     item.Headline,
		Bio:          item.Bio,
		Expertise:    item.Expertise,
		Availability: textValue(item.Availability),
		AvatarURL:    textValue(item.AvatarUrl),
	}, nil
}

func mapSession(item queries.MentorshipSession) (SessionRequestItem, error) {
	id, err := uuidFromPg(item.ID)
	if err != nil {
		return SessionRequestItem{}, fmt.Errorf("mentorship: session id: %w", err)
	}
	mentorID, err := uuidFromPg(item.MentorID)
	if err != nil {
		return SessionRequestItem{}, fmt.Errorf("mentorship: session mentor id: %w", err)
	}
	requesterID, err := uuidFromPg(item.RequesterID)
	if err != nil {
		return SessionRequestItem{}, fmt.Errorf("mentorship: session requester id: %w", err)
	}
	return SessionRequestItem{
		ID:          id.String(),
		MentorID:    mentorID.String(),
		RequesterID: requesterID.String(),
		Message:     item.Message,
		Status:      item.Status,
		ScheduledAt: timestamptzValue(item.ScheduledAt),
	}, nil
}

func uuidToPg(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
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
	return value.Time.UTC().Format(time.RFC3339)
}
