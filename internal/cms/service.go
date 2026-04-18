package cms

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var slugUnsafePattern = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct{ repo Repository }

type CreateArticleInput struct {
	ID             string
	AuthorID       string
	Title          string
	Excerpt        string
	Content        string
	CoverImageURL  string
	Type           string
	SourceName     string
	SourceURL      string
	SEOTitle       string
	SEODescription string
	IsFeatured     *bool
}

type CreateOpportunityInput struct {
	ID           string
	PublishedBy  string
	Title        string
	Type         string
	EntityName   string
	Description  string
	Requirements string
	Deadline     string
	ApplyURL     string
	Country      string
	Language     string
	Area         string
}

type ArticleResult struct {
	Data ArticleItem `json:"data"`
}
type OpportunityResult struct {
	Data OpportunityItem `json:"data"`
}

type ArticleItem struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	AuthorID  string `json:"author_id"`
	Published string `json:"published_at,omitempty"`
}

type OpportunityItem struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	EntityName  string `json:"entity_name"`
	Verified    bool   `json:"verified"`
	IsActive    bool   `json:"is_active"`
	PublishedBy string `json:"published_by"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) CreateArticle(ctx context.Context, input CreateArticleInput) (*ArticleResult, error) {
	authorID, err := uuid.Parse(strings.TrimSpace(input.AuthorID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid author id: %w", err)
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Content) == "" || strings.TrimSpace(input.Type) == "" {
		return nil, fmt.Errorf("cms: title, content and type are required")
	}
	article, err := s.repo.CreateArticle(ctx, queries.CreateArticleParams{
		Slug:           slugify(input.Title),
		Title:          strings.TrimSpace(input.Title),
		Excerpt:        textToPg(input.Excerpt),
		Content:        strings.TrimSpace(input.Content),
		CoverImageUrl:  textToPg(input.CoverImageURL),
		Type:           strings.TrimSpace(input.Type),
		Status:         "draft",
		SourceName:     textToPg(input.SourceName),
		SourceUrl:      textToPg(input.SourceURL),
		SeoTitle:       textToPg(input.SEOTitle),
		SeoDescription: textToPg(input.SEODescription),
		IsFeatured:     boolValue(input.IsFeatured, false),
		AuthorID:       uuidToPg(authorID),
		PublishedAt:    pgtype.Timestamptz{},
	})
	if err != nil {
		return nil, fmt.Errorf("cms: create article: %w", err)
	}
	mapped, err := mapArticle(article)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

func (s *Service) UpdateArticle(ctx context.Context, input CreateArticleInput) (*ArticleResult, error) {
	articleID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid article id: %w", err)
	}
	existingArticle, err := s.repo.GetArticleByID(ctx, uuidToPg(articleID))
	if err != nil {
		return nil, fmt.Errorf("cms: get article: %w", err)
	}
	title := chooseString(input.Title, existingArticle.Title)
	content := chooseString(input.Content, existingArticle.Content)
	articleType := chooseString(input.Type, existingArticle.Type)
	if strings.TrimSpace(title) == "" || strings.TrimSpace(content) == "" || strings.TrimSpace(articleType) == "" {
		return nil, fmt.Errorf("cms: title, content and type are required")
	}
	updatedArticle, err := s.repo.UpdateArticle(ctx, queries.UpdateArticleParams{
		ID:             uuidToPg(articleID),
		Title:          title,
		Excerpt:        chooseText(input.Excerpt, existingArticle.Excerpt),
		Content:        content,
		CoverImageUrl:  chooseText(input.CoverImageURL, existingArticle.CoverImageUrl),
		Type:           articleType,
		SourceName:     chooseText(input.SourceName, existingArticle.SourceName),
		SourceUrl:      chooseText(input.SourceURL, existingArticle.SourceUrl),
		SeoTitle:       chooseText(input.SEOTitle, existingArticle.SeoTitle),
		SeoDescription: chooseText(input.SEODescription, existingArticle.SeoDescription),
		IsFeatured:     boolValue(input.IsFeatured, existingArticle.IsFeatured),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: update article: %w", err)
	}
	mapped, err := mapArticle(updatedArticle)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

func (s *Service) CreateOpportunity(ctx context.Context, input CreateOpportunityInput) (*OpportunityResult, error) {
	publishedBy, err := uuid.Parse(strings.TrimSpace(input.PublishedBy))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid publisher id: %w", err)
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Type) == "" || strings.TrimSpace(input.EntityName) == "" || strings.TrimSpace(input.Description) == "" || strings.TrimSpace(input.Country) == "" {
		return nil, fmt.Errorf("cms: required fields are missing")
	}
	var deadline pgtype.Timestamptz
	if strings.TrimSpace(input.Deadline) != "" {
		parsedTime, err := time.Parse(time.RFC3339, strings.TrimSpace(input.Deadline))
		if err != nil {
			return nil, fmt.Errorf("cms: invalid deadline: %w", err)
		}
		deadline = pgtype.Timestamptz{Time: parsedTime, Valid: true}
	}
	opportunity, err := s.repo.CreateOpportunity(ctx, queries.CreateOpportunityParams{
		Slug:         slugify(input.Title),
		Title:        strings.TrimSpace(input.Title),
		Type:         strings.TrimSpace(input.Type),
		EntityName:   strings.TrimSpace(input.EntityName),
		Description:  strings.TrimSpace(input.Description),
		Requirements: textToPg(input.Requirements),
		Deadline:     deadline,
		ApplyUrl:     textToPg(input.ApplyURL),
		Country:      strings.TrimSpace(input.Country),
		Language:     textToPg(input.Language),
		Area:         textToPg(input.Area),
		IsActive:     false,
		PublishedBy:  uuidToPg(publishedBy),
		Verified:     false,
	})
	if err != nil {
		return nil, fmt.Errorf("cms: create opportunity: %w", err)
	}
	mapped, err := mapOpportunity(opportunity)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func (s *Service) UpdateOpportunity(ctx context.Context, input CreateOpportunityInput) (*OpportunityResult, error) {
	opportunityID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid opportunity id: %w", err)
	}
	existingOpportunity, err := s.repo.GetOpportunityByID(ctx, uuidToPg(opportunityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get opportunity: %w", err)
	}
	title := chooseString(input.Title, existingOpportunity.Title)
	opportunityType := chooseString(input.Type, existingOpportunity.Type)
	entityName := chooseString(input.EntityName, existingOpportunity.EntityName)
	description := chooseString(input.Description, existingOpportunity.Description)
	country := chooseString(input.Country, existingOpportunity.Country)
	if title == "" || opportunityType == "" || entityName == "" || description == "" || country == "" {
		return nil, fmt.Errorf("cms: required fields are missing")
	}
	deadline := existingOpportunity.Deadline
	if strings.TrimSpace(input.Deadline) != "" {
		parsedTime, err := time.Parse(time.RFC3339, strings.TrimSpace(input.Deadline))
		if err != nil {
			return nil, fmt.Errorf("cms: invalid deadline: %w", err)
		}
		deadline = pgtype.Timestamptz{Time: parsedTime, Valid: true}
	}
	updatedOpportunity, err := s.repo.UpdateOpportunity(ctx, queries.UpdateOpportunityParams{
		ID:           uuidToPg(opportunityID),
		Title:        title,
		Type:         opportunityType,
		EntityName:   entityName,
		Description:  description,
		Requirements: chooseText(input.Requirements, existingOpportunity.Requirements),
		Deadline:     deadline,
		ApplyUrl:     chooseText(input.ApplyURL, existingOpportunity.ApplyUrl),
		Country:      country,
		Language:     chooseText(input.Language, existingOpportunity.Language),
		Area:         chooseText(input.Area, existingOpportunity.Area),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: update opportunity: %w", err)
	}
	mapped, err := mapOpportunity(updatedOpportunity)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func mapArticle(article queries.Article) (ArticleItem, error) {
	id, err := uuidFromPg(article.ID)
	if err != nil {
		return ArticleItem{}, fmt.Errorf("cms: article id: %w", err)
	}
	authorID, err := uuidFromPg(article.AuthorID)
	if err != nil {
		return ArticleItem{}, fmt.Errorf("cms: article author id: %w", err)
	}
	return ArticleItem{ID: id.String(), Slug: article.Slug, Title: article.Title, Type: article.Type, Status: article.Status, AuthorID: authorID.String(), Published: timestamptzValue(article.PublishedAt)}, nil
}

func mapOpportunity(opportunity queries.Opportunity) (OpportunityItem, error) {
	id, err := uuidFromPg(opportunity.ID)
	if err != nil {
		return OpportunityItem{}, fmt.Errorf("cms: opportunity id: %w", err)
	}
	publishedBy, err := uuidFromPg(opportunity.PublishedBy)
	if err != nil {
		return OpportunityItem{}, fmt.Errorf("cms: opportunity publisher id: %w", err)
	}
	return OpportunityItem{ID: id.String(), Slug: opportunity.Slug, Title: opportunity.Title, Type: opportunity.Type, EntityName: opportunity.EntityName, Verified: opportunity.Verified, IsActive: opportunity.IsActive, PublishedBy: publishedBy.String()}, nil
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = slugUnsafePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return uuid.NewString()
	}
	return value
}

func chooseString(input string, fallback string) string {
	if strings.TrimSpace(input) == "" {
		return fallback
	}
	return strings.TrimSpace(input)
}

func chooseText(input string, fallback pgtype.Text) pgtype.Text {
	if strings.TrimSpace(input) == "" {
		return fallback
	}
	return textToPg(input)
}

func boolValue(input *bool, fallback bool) bool {
	if input == nil {
		return fallback
	}
	return *input
}

func uuidToPg(value uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte(value), Valid: true} }
func uuidFromPg(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, fmt.Errorf("invalid uuid")
	}
	return uuid.UUID(value.Bytes), nil
}
func textToPg(value string) pgtype.Text {
	value = strings.TrimSpace(value)
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
func timestamptzValue(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
