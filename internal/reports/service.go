package reports

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

type CreateReportInput struct {
	ReporterID string
	EntityType string
	EntityID   string
	Reason     string
}

type Result struct {
	Data Item `json:"data"`
}

type Item struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
	ResolvedAt string `json:"resolved_at,omitempty"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) Create(ctx context.Context, input CreateReportInput) (*Result, error) {
	reporterID, err := uuid.Parse(strings.TrimSpace(input.ReporterID))
	if err != nil {
		return nil, fmt.Errorf("reports: invalid reporter id: %w", err)
	}
	entityID, err := uuid.Parse(strings.TrimSpace(input.EntityID))
	if err != nil {
		return nil, fmt.Errorf("reports: invalid entity id: %w", err)
	}
	exists, err := s.entityExists(ctx, strings.TrimSpace(input.EntityType), pgtype.UUID{Bytes: [16]byte(entityID), Valid: true})
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("reports: entity does not exist")
	}
	item, err := s.repo.CreateReport(ctx, queries.CreateReportParams{
		ReporterID: pgtype.UUID{Bytes: [16]byte(reporterID), Valid: true},
		EntityType: strings.TrimSpace(input.EntityType),
		EntityID:   pgtype.UUID{Bytes: [16]byte(entityID), Valid: true},
		Reason:     strings.TrimSpace(input.Reason),
	})
	if err != nil {
		return nil, fmt.Errorf("reports: create report: %w", err)
	}
	mapped, err := mapReport(item)
	if err != nil {
		return nil, err
	}
	return &Result{Data: mapped}, nil
}

func (s *Service) entityExists(ctx context.Context, entityType string, entityID pgtype.UUID) (bool, error) {
	switch entityType {
	case "university":
		exists, err := s.repo.ReportUniversityExists(ctx, entityID)
		if err != nil { return false, fmt.Errorf("reports: check university exists: %w", err) }
		return exists, nil
	case "course":
		exists, err := s.repo.ReportCourseExists(ctx, entityID)
		if err != nil { return false, fmt.Errorf("reports: check course exists: %w", err) }
		return exists, nil
	case "opportunity":
		exists, err := s.repo.ReportOpportunityExists(ctx, entityID)
		if err != nil { return false, fmt.Errorf("reports: check opportunity exists: %w", err) }
		return exists, nil
	default:
		return false, fmt.Errorf("reports: invalid entity type")
	}
}

func mapReport(report queries.Report) (Item, error) {
	id, err := uuidFromPg(report.ID)
	if err != nil {
		return Item{}, fmt.Errorf("reports: id: %w", err)
	}
	reporterID, err := uuidFromPg(report.ReporterID)
	if err != nil {
		return Item{}, fmt.Errorf("reports: reporter id: %w", err)
	}
	entityID, err := uuidFromPg(report.EntityID)
	if err != nil {
		return Item{}, fmt.Errorf("reports: entity id: %w", err)
	}
	return Item{
		ID:         id.String(),
		ReporterID: reporterID.String(),
		EntityType: report.EntityType,
		EntityID:   entityID.String(),
		Reason:     report.Reason,
		Status:     report.Status,
		ResolvedAt: timestamptzValue(report.ResolvedAt),
	}, nil
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
