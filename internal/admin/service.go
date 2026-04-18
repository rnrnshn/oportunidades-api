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

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

type ArticleResult struct {
	Data ArticleItem `json:"data"`
}
type OpportunityResult struct {
	Data OpportunityItem `json:"data"`
}

type ReportsResult struct {
	Data []ReportItem   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type ReportResult struct {
	Data ReportItem `json:"data"`
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

type UpdateReportStatusInput struct {
	ReportID string
	Status   string
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

type ReportItem struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Reason     string `json:"reason"`
	Status     string `json:"status"`
	ResolvedAt string `json:"resolved_at,omitempty"`
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

func (s *Service) UnpublishArticle(ctx context.Context, articleID string) (*ArticleResult, error) {
	id, err := parseID(articleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetArticleByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.UnpublishArticle(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := mapArticle(item)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

func (s *Service) ArchiveArticle(ctx context.Context, articleID string) (*ArticleResult, error) {
	id, err := parseID(articleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetArticleByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.ArchiveArticle(ctx, id)
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

func (s *Service) RejectOpportunity(ctx context.Context, opportunityID string) (*OpportunityResult, error) {
	id, err := parseID(opportunityID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetOpportunityByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.RejectOpportunity(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := mapOpportunity(item)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func (s *Service) DeactivateOpportunity(ctx context.Context, opportunityID string) (*OpportunityResult, error) {
	id, err := parseID(opportunityID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetOpportunityByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.DeactivateOpportunity(ctx, id)
	if err != nil {
		return nil, err
	}
	mapped, err := mapOpportunity(item)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func (s *Service) ListReports(ctx context.Context, params PaginationParams) (*ReportsResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListReports(ctx, queries.ListReportsParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)})
	if err != nil {
		return nil, fmt.Errorf("admin: list reports: %w", err)
	}
	total, err := s.repo.CountReports(ctx)
	if err != nil {
		return nil, fmt.Errorf("admin: count reports: %w", err)
	}
	data := make([]ReportItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapReport(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mapped)
	}
	return &ReportsResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) UpdateReportStatus(ctx context.Context, input UpdateReportStatusInput) (*ReportResult, error) {
	id, err := parseID(input.ReportID)
	if err != nil {
		return nil, ErrNotFound
	}
	if _, err := s.repo.GetReportByID(ctx, id); err != nil {
		return nil, err
	}
	item, err := s.repo.UpdateReportStatus(ctx, queries.UpdateReportStatusParams{ID: id, Status: strings.TrimSpace(input.Status)})
	if err != nil {
		return nil, err
	}
	mapped, err := mapReport(item)
	if err != nil {
		return nil, err
	}
	return &ReportResult{Data: mapped}, nil
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

func mapReport(report queries.Report) (ReportItem, error) {
	id, err := uuidFromPg(report.ID)
	if err != nil {
		return ReportItem{}, fmt.Errorf("admin: report id: %w", err)
	}
	reporterID, err := uuidFromPg(report.ReporterID)
	if err != nil {
		return ReportItem{}, fmt.Errorf("admin: report reporter id: %w", err)
	}
	entityID, err := uuidFromPg(report.EntityID)
	if err != nil {
		return ReportItem{}, fmt.Errorf("admin: report entity id: %w", err)
	}
	return ReportItem{ID: id.String(), ReporterID: reporterID.String(), EntityType: report.EntityType, EntityID: entityID.String(), Reason: report.Reason, Status: report.Status, ResolvedAt: timestamptzValue(report.ResolvedAt)}, nil
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
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	return PaginationMeta{Total: total, Page: page, PerPage: perPage, TotalPages: totalPages}
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
