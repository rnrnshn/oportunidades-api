package cms

import (
	"context"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

var slugUnsafePattern = regexp.MustCompile(`[^a-z0-9]+`)

const (
	defaultPage    = 1
	defaultPerPage = 20
	maxPerPage     = 100
)

type Service struct{ repo Repository }

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

type ArticlesResult struct {
	Data []ArticleItem  `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type OpportunitiesResult struct {
	Data []OpportunityItem `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

type UniversitiesResult struct {
	Data []UniversityItem `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

type CoursesResult struct {
	Data []CourseItem   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

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

type CreateUniversityInput struct {
	ID          string
	CreatedBy   string
	Name        string
	Type        string
	Province    string
	Description string
	LogoURL     string
	Website     string
	Email       string
	Phone       string
}

type CreateCourseInput struct {
	ID                string
	UniversityID      string
	Name              string
	Area              string
	Level             string
	Regime            string
	DurationYears     int32
	HasDurationYears  bool
	AnnualFee         string
	EntryRequirements string
}

type ArticleResult struct {
	Data ArticleItem `json:"data"`
}
type OpportunityResult struct {
	Data OpportunityItem `json:"data"`
}

type UniversityResult struct {
	Data UniversityItem `json:"data"`
}

type CourseResult struct {
	Data CourseItem `json:"data"`
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

type UniversityItem struct {
	ID        string `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Province  string `json:"province"`
	Verified  bool   `json:"verified"`
	CreatedBy string `json:"created_by"`
}

type CourseItem struct {
	ID           string `json:"id"`
	Slug         string `json:"slug"`
	UniversityID string `json:"university_id"`
	Name         string `json:"name"`
	Area         string `json:"area"`
	Level        string `json:"level"`
	Regime       string `json:"regime"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListArticles(ctx context.Context, params PaginationParams) (*ArticlesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListCMSArticles(ctx, queries.ListCMSArticlesParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)})
	if err != nil {
		return nil, fmt.Errorf("cms: list articles: %w", err)
	}
	total, err := s.repo.CountCMSArticles(ctx)
	if err != nil {
		return nil, fmt.Errorf("cms: count articles: %w", err)
	}
	data := make([]ArticleItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapArticle(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mapped)
	}
	return &ArticlesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetArticle(ctx context.Context, id string) (*ArticleResult, error) {
	articleID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid article id: %w", err)
	}
	article, err := s.repo.GetArticleByID(ctx, uuidToPg(articleID))
	if err != nil {
		return nil, fmt.Errorf("cms: get article: %w", err)
	}
	mapped, err := mapArticle(article)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

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

func (s *Service) ListUniversities(ctx context.Context, params PaginationParams) (*UniversitiesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListCMSUniversities(ctx, queries.ListCMSUniversitiesParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)})
	if err != nil {
		return nil, fmt.Errorf("cms: list universities: %w", err)
	}
	total, err := s.repo.CountCMSUniversities(ctx)
	if err != nil {
		return nil, fmt.Errorf("cms: count universities: %w", err)
	}
	data := make([]UniversityItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapUniversity(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mapped)
	}
	return &UniversitiesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetUniversity(ctx context.Context, id string) (*UniversityResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	item, err := s.repo.GetUniversityByID(ctx, uuidToPg(universityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get university: %w", err)
	}
	mapped, err := mapUniversity(item)
	if err != nil {
		return nil, err
	}
	return &UniversityResult{Data: mapped}, nil
}

func (s *Service) CreateUniversity(ctx context.Context, input CreateUniversityInput) (*UniversityResult, error) {
	createdBy, err := uuid.Parse(strings.TrimSpace(input.CreatedBy))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid creator id: %w", err)
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Type) == "" || strings.TrimSpace(input.Province) == "" {
		return nil, fmt.Errorf("cms: university required fields are missing")
	}
	item, err := s.repo.CreateUniversity(ctx, queries.CreateUniversityParams{
		Slug:        slugify(input.Name),
		Name:        strings.TrimSpace(input.Name),
		Type:        strings.TrimSpace(input.Type),
		Province:    strings.TrimSpace(input.Province),
		Description: textToPg(input.Description),
		LogoUrl:     textToPg(input.LogoURL),
		Website:     textToPg(input.Website),
		Email:       textToPg(input.Email),
		Phone:       textToPg(input.Phone),
		Verified:    false,
		VerifiedAt:  pgtype.Timestamptz{},
		CreatedBy:   uuidToPg(createdBy),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: create university: %w", err)
	}
	mapped, err := mapUniversity(item)
	if err != nil {
		return nil, err
	}
	return &UniversityResult{Data: mapped}, nil
}

func (s *Service) UpdateUniversity(ctx context.Context, input CreateUniversityInput) (*UniversityResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	existing, err := s.repo.GetUniversityByID(ctx, uuidToPg(universityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get university: %w", err)
	}
	name := chooseString(input.Name, existing.Name)
	typeValue := chooseString(input.Type, existing.Type)
	province := chooseString(input.Province, existing.Province)
	if name == "" || typeValue == "" || province == "" {
		return nil, fmt.Errorf("cms: university required fields are missing")
	}
	item, err := s.repo.UpdateUniversity(ctx, queries.UpdateUniversityParams{
		ID:          uuidToPg(universityID),
		Name:        name,
		Type:        typeValue,
		Province:    province,
		Description: chooseText(input.Description, existing.Description),
		LogoUrl:     chooseText(input.LogoURL, existing.LogoUrl),
		Website:     chooseText(input.Website, existing.Website),
		Email:       chooseText(input.Email, existing.Email),
		Phone:       chooseText(input.Phone, existing.Phone),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: update university: %w", err)
	}
	mapped, err := mapUniversity(item)
	if err != nil {
		return nil, err
	}
	return &UniversityResult{Data: mapped}, nil
}

func (s *Service) ListCourses(ctx context.Context, params PaginationParams) (*CoursesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListCMSCourses(ctx, queries.ListCMSCoursesParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)})
	if err != nil {
		return nil, fmt.Errorf("cms: list courses: %w", err)
	}
	total, err := s.repo.CountCMSCourses(ctx)
	if err != nil {
		return nil, fmt.Errorf("cms: count courses: %w", err)
	}
	data := make([]CourseItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapCourse(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mapped)
	}
	return &CoursesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetCourse(ctx context.Context, id string) (*CourseResult, error) {
	courseID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid course id: %w", err)
	}
	item, err := s.repo.GetCourseByID(ctx, uuidToPg(courseID))
	if err != nil {
		return nil, fmt.Errorf("cms: get course: %w", err)
	}
	mapped, err := mapCourse(item)
	if err != nil {
		return nil, err
	}
	return &CourseResult{Data: mapped}, nil
}

func (s *Service) CreateCourse(ctx context.Context, input CreateCourseInput) (*CourseResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(input.UniversityID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Area) == "" || strings.TrimSpace(input.Level) == "" || strings.TrimSpace(input.Regime) == "" {
		return nil, fmt.Errorf("cms: course required fields are missing")
	}
	item, err := s.repo.CreateCourse(ctx, queries.CreateCourseParams{
		Slug:              slugify(input.Name),
		UniversityID:      uuidToPg(universityID),
		Name:              strings.TrimSpace(input.Name),
		Area:              strings.TrimSpace(input.Area),
		Level:             strings.TrimSpace(input.Level),
		Regime:            strings.TrimSpace(input.Regime),
		DurationYears:     chooseInt4(input.DurationYears, input.HasDurationYears),
		AnnualFee:         numericToPg(input.AnnualFee),
		EntryRequirements: textToPg(input.EntryRequirements),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: create course: %w", err)
	}
	mapped, err := mapCourse(item)
	if err != nil {
		return nil, err
	}
	return &CourseResult{Data: mapped}, nil
}

func (s *Service) UpdateCourse(ctx context.Context, input CreateCourseInput) (*CourseResult, error) {
	courseID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid course id: %w", err)
	}
	existing, err := s.repo.GetCourseByID(ctx, uuidToPg(courseID))
	if err != nil {
		return nil, fmt.Errorf("cms: get course: %w", err)
	}
	universityID := existing.UniversityID
	if strings.TrimSpace(input.UniversityID) != "" {
		parsedUniversityID, err := uuid.Parse(strings.TrimSpace(input.UniversityID))
		if err != nil {
			return nil, fmt.Errorf("cms: invalid university id: %w", err)
		}
		universityID = uuidToPg(parsedUniversityID)
	}
	name := chooseString(input.Name, existing.Name)
	area := chooseString(input.Area, existing.Area)
	level := chooseString(input.Level, existing.Level)
	regime := chooseString(input.Regime, existing.Regime)
	if name == "" || area == "" || level == "" || regime == "" {
		return nil, fmt.Errorf("cms: course required fields are missing")
	}
	durationYears := existing.DurationYears
	if input.HasDurationYears {
		durationYears = chooseInt4(input.DurationYears, true)
	}
	annualFee := existing.AnnualFee
	if strings.TrimSpace(input.AnnualFee) != "" {
		annualFee = numericToPg(input.AnnualFee)
	}
	item, err := s.repo.UpdateCourse(ctx, queries.UpdateCourseParams{
		ID:                uuidToPg(courseID),
		UniversityID:      universityID,
		Name:              name,
		Area:              area,
		Level:             level,
		Regime:            regime,
		DurationYears:     durationYears,
		AnnualFee:         annualFee,
		EntryRequirements: chooseText(input.EntryRequirements, existing.EntryRequirements),
	})
	if err != nil {
		return nil, fmt.Errorf("cms: update course: %w", err)
	}
	mapped, err := mapCourse(item)
	if err != nil {
		return nil, err
	}
	return &CourseResult{Data: mapped}, nil
}

func (s *Service) ListOpportunities(ctx context.Context, params PaginationParams) (*OpportunitiesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListCMSOpportunities(ctx, queries.ListCMSOpportunitiesParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)})
	if err != nil {
		return nil, fmt.Errorf("cms: list opportunities: %w", err)
	}
	total, err := s.repo.CountCMSOpportunities(ctx)
	if err != nil {
		return nil, fmt.Errorf("cms: count opportunities: %w", err)
	}
	data := make([]OpportunityItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapOpportunity(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mapped)
	}
	return &OpportunitiesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetOpportunity(ctx context.Context, id string) (*OpportunityResult, error) {
	opportunityID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid opportunity id: %w", err)
	}
	opportunity, err := s.repo.GetOpportunityByID(ctx, uuidToPg(opportunityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get opportunity: %w", err)
	}
	mapped, err := mapOpportunity(opportunity)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
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

func mapUniversity(university queries.University) (UniversityItem, error) {
	id, err := uuidFromPg(university.ID)
	if err != nil {
		return UniversityItem{}, fmt.Errorf("cms: university id: %w", err)
	}
	createdBy, err := uuidFromPg(university.CreatedBy)
	if err != nil {
		return UniversityItem{}, fmt.Errorf("cms: university creator id: %w", err)
	}
	return UniversityItem{ID: id.String(), Slug: university.Slug, Name: university.Name, Type: university.Type, Province: university.Province, Verified: university.Verified, CreatedBy: createdBy.String()}, nil
}

func mapCourse(course queries.Course) (CourseItem, error) {
	id, err := uuidFromPg(course.ID)
	if err != nil {
		return CourseItem{}, fmt.Errorf("cms: course id: %w", err)
	}
	universityID, err := uuidFromPg(course.UniversityID)
	if err != nil {
		return CourseItem{}, fmt.Errorf("cms: course university id: %w", err)
	}
	return CourseItem{ID: id.String(), Slug: course.Slug, UniversityID: universityID.String(), Name: course.Name, Area: course.Area, Level: course.Level, Regime: course.Regime}, nil
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

func chooseInt4(value int32, valid bool) pgtype.Int4 {
	if !valid {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: value, Valid: true}
}

func numericToPg(value string) pgtype.Numeric {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Numeric{}
	}
	return pgtype.Numeric{Int: mustParseNumeric(trimmed), Exp: -2, Valid: true}
}

func mustParseNumeric(value string) *big.Int {
	clean := strings.ReplaceAll(strings.TrimSpace(value), ".", "")
	parsed := new(big.Int)
	parsed.SetString(clean, 10)
	return parsed
}
func timestamptzValue(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.UTC().Format(time.RFC3339)
}
