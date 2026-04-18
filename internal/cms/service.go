package cms

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"sort"
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

type Actor struct {
	UserID string
	Role   string
}

type PaginationParams struct {
	Page    int
	PerPage int
}

type ArticleListFilters struct {
	Query    string
	Type     string
	Status   string
	Featured *bool
	Sort     string
}

type OpportunityListFilters struct {
	Query    string
	Type     string
	Verified *bool
	Active   *bool
	Sort     string
}

type UniversityListFilters struct {
	Query    string
	Type     string
	Province string
	Verified *bool
	Sort     string
}

type CourseListFilters struct {
	Query        string
	Area         string
	Level        string
	Regime       string
	UniversityID string
	Sort         string
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
	ContentJSON    json.RawMessage
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
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt,omitempty"`
	Content        string `json:"content"`
	ContentJSON    any    `json:"content_json,omitempty"`
	CoverImageURL  string `json:"cover_image_url,omitempty"`
	Type           string `json:"type"`
	Status         string `json:"status"`
	SourceName     string `json:"source_name,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	IsFeatured     bool   `json:"is_featured"`
	AuthorID       string `json:"author_id"`
	Published      string `json:"published_at,omitempty"`
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

func (s *Service) ListArticles(ctx context.Context, actor Actor, params PaginationParams, filters ArticleListFilters) (*ArticlesResult, error) {
	page, perPage := normalizePagination(params)
	totalBeforeFilter, err := s.repo.CountCMSArticles(ctx, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: count articles: %w", err)
	}
	limit := totalBeforeFilter
	if limit < 1 {
		limit = 1
	}
	items, err := s.repo.ListCMSArticles(ctx, queries.ListCMSArticlesParams{Limit: int32(limit), Offset: 0}, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: list articles: %w", err)
	}
	items = filterArticles(items, filters)
	items, _, err = s.filterArticlesByActor(items, int64(len(items)), actor, params)
	if err != nil {
		return nil, err
	}
	sortArticles(items, filters.Sort)
	total := int64(len(items))
	items = paginateArticles(items, page, perPage)
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

func (s *Service) GetArticle(ctx context.Context, actor Actor, id string) (*ArticleResult, error) {
	articleID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid article id: %w", err)
	}
	article, err := s.repo.GetArticleByID(ctx, uuidToPg(articleID))
	if err != nil {
		return nil, fmt.Errorf("cms: get article: %w", err)
	}
	if !canManageArticle(actor, article) {
		return nil, ErrForbidden
	}
	mapped, err := mapArticle(article)
	if err != nil {
		return nil, err
	}
	return &ArticleResult{Data: mapped}, nil
}

func (s *Service) CreateArticle(ctx context.Context, actor Actor, input CreateArticleInput) (*ArticleResult, error) {
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
		ContentJson:    jsonBytes(input.ContentJSON),
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

func (s *Service) UpdateArticle(ctx context.Context, actor Actor, input CreateArticleInput) (*ArticleResult, error) {
	articleID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid article id: %w", err)
	}
	existingArticle, err := s.repo.GetArticleByID(ctx, uuidToPg(articleID))
	if err != nil {
		return nil, fmt.Errorf("cms: get article: %w", err)
	}
	if !canManageArticle(actor, existingArticle) {
		return nil, ErrForbidden
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
		ContentJson:    chooseJSON(input.ContentJSON, existingArticle.ContentJson),
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

func (s *Service) ListUniversities(ctx context.Context, actor Actor, params PaginationParams, filters UniversityListFilters) (*UniversitiesResult, error) {
	page, perPage := normalizePagination(params)
	totalBeforeFilter, err := s.repo.CountCMSUniversities(ctx, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: count universities: %w", err)
	}
	limit := totalBeforeFilter
	if limit < 1 {
		limit = 1
	}
	items, err := s.repo.ListCMSUniversities(ctx, queries.ListCMSUniversitiesParams{Limit: int32(limit), Offset: 0}, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: list universities: %w", err)
	}
	items = filterUniversities(items, filters)
	items, _, err = s.filterUniversitiesByActor(items, int64(len(items)), actor, params)
	if err != nil {
		return nil, err
	}
	sortUniversities(items, filters.Sort)
	total := int64(len(items))
	items = paginateUniversities(items, page, perPage)
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

func (s *Service) GetUniversity(ctx context.Context, actor Actor, id string) (*UniversityResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	item, err := s.repo.GetUniversityByID(ctx, uuidToPg(universityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get university: %w", err)
	}
	if !canManageUniversity(actor, item) {
		return nil, ErrForbidden
	}
	mapped, err := mapUniversity(item)
	if err != nil {
		return nil, err
	}
	return &UniversityResult{Data: mapped}, nil
}

func (s *Service) CreateUniversity(ctx context.Context, actor Actor, input CreateUniversityInput) (*UniversityResult, error) {
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

func (s *Service) UpdateUniversity(ctx context.Context, actor Actor, input CreateUniversityInput) (*UniversityResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	existing, err := s.repo.GetUniversityByID(ctx, uuidToPg(universityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get university: %w", err)
	}
	if !canManageUniversity(actor, existing) {
		return nil, ErrForbidden
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

func (s *Service) ListCourses(ctx context.Context, actor Actor, params PaginationParams, filters CourseListFilters) (*CoursesResult, error) {
	page, perPage := normalizePagination(params)
	totalBeforeFilter, err := s.repo.CountCMSCourses(ctx, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: count courses: %w", err)
	}
	limit := totalBeforeFilter
	if limit < 1 {
		limit = 1
	}
	items, err := s.repo.ListCMSCourses(ctx, queries.ListCMSCoursesParams{Limit: int32(limit), Offset: 0}, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: list courses: %w", err)
	}
	items = filterCourses(items, filters)
	items, _, err = s.filterCoursesByActor(ctx, items, int64(len(items)), actor, params)
	if err != nil {
		return nil, err
	}
	sortCourses(items, filters.Sort)
	total := int64(len(items))
	items = paginateCourses(items, page, perPage)
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

func (s *Service) GetCourse(ctx context.Context, actor Actor, id string) (*CourseResult, error) {
	courseID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid course id: %w", err)
	}
	item, err := s.repo.GetCourseByID(ctx, uuidToPg(courseID))
	if err != nil {
		return nil, fmt.Errorf("cms: get course: %w", err)
	}
	ok, err := s.canManageCourse(ctx, actor, item)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	mapped, err := mapCourse(item)
	if err != nil {
		return nil, err
	}
	return &CourseResult{Data: mapped}, nil
}

func (s *Service) CreateCourse(ctx context.Context, actor Actor, input CreateCourseInput) (*CourseResult, error) {
	universityID, err := uuid.Parse(strings.TrimSpace(input.UniversityID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid university id: %w", err)
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Area) == "" || strings.TrimSpace(input.Level) == "" || strings.TrimSpace(input.Regime) == "" {
		return nil, fmt.Errorf("cms: course required fields are missing")
	}
	university, err := s.repo.GetUniversityByID(ctx, uuidToPg(universityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get university for course: %w", err)
	}
	if !canManageUniversity(actor, university) {
		return nil, ErrForbidden
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

func (s *Service) UpdateCourse(ctx context.Context, actor Actor, input CreateCourseInput) (*CourseResult, error) {
	courseID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid course id: %w", err)
	}
	existing, err := s.repo.GetCourseByID(ctx, uuidToPg(courseID))
	if err != nil {
		return nil, fmt.Errorf("cms: get course: %w", err)
	}
	ok, err := s.canManageCourse(ctx, actor, existing)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	universityID := existing.UniversityID
	if strings.TrimSpace(input.UniversityID) != "" {
		parsedUniversityID, err := uuid.Parse(strings.TrimSpace(input.UniversityID))
		if err != nil {
			return nil, fmt.Errorf("cms: invalid university id: %w", err)
		}
		universityID = uuidToPg(parsedUniversityID)
		university, err := s.repo.GetUniversityByID(ctx, universityID)
		if err != nil {
			return nil, fmt.Errorf("cms: get university for course update: %w", err)
		}
		if !canManageUniversity(actor, university) {
			return nil, ErrForbidden
		}
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

func (s *Service) ListOpportunities(ctx context.Context, actor Actor, params PaginationParams, filters OpportunityListFilters) (*OpportunitiesResult, error) {
	page, perPage := normalizePagination(params)
	totalBeforeFilter, err := s.repo.CountCMSOpportunities(ctx, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: count opportunities: %w", err)
	}
	limit := totalBeforeFilter
	if limit < 1 {
		limit = 1
	}
	items, err := s.repo.ListCMSOpportunities(ctx, queries.ListCMSOpportunitiesParams{Limit: int32(limit), Offset: 0}, actor, filters)
	if err != nil {
		return nil, fmt.Errorf("cms: list opportunities: %w", err)
	}
	items = filterOpportunities(items, filters)
	items, _, err = s.filterOpportunitiesByActor(items, int64(len(items)), actor, params)
	if err != nil {
		return nil, err
	}
	sortOpportunities(items, filters.Sort)
	total := int64(len(items))
	items = paginateOpportunities(items, page, perPage)
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

func (s *Service) GetOpportunity(ctx context.Context, actor Actor, id string) (*OpportunityResult, error) {
	opportunityID, err := uuid.Parse(strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid opportunity id: %w", err)
	}
	opportunity, err := s.repo.GetOpportunityByID(ctx, uuidToPg(opportunityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get opportunity: %w", err)
	}
	if !canManageOpportunity(actor, opportunity) {
		return nil, ErrForbidden
	}
	mapped, err := mapOpportunity(opportunity)
	if err != nil {
		return nil, err
	}
	return &OpportunityResult{Data: mapped}, nil
}

func (s *Service) CreateOpportunity(ctx context.Context, actor Actor, input CreateOpportunityInput) (*OpportunityResult, error) {
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

func (s *Service) UpdateOpportunity(ctx context.Context, actor Actor, input CreateOpportunityInput) (*OpportunityResult, error) {
	opportunityID, err := uuid.Parse(strings.TrimSpace(input.ID))
	if err != nil {
		return nil, fmt.Errorf("cms: invalid opportunity id: %w", err)
	}
	existingOpportunity, err := s.repo.GetOpportunityByID(ctx, uuidToPg(opportunityID))
	if err != nil {
		return nil, fmt.Errorf("cms: get opportunity: %w", err)
	}
	if !canManageOpportunity(actor, existingOpportunity) {
		return nil, ErrForbidden
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
	return ArticleItem{
		ID:             id.String(),
		Slug:           article.Slug,
		Title:          article.Title,
		Excerpt:        textValue(article.Excerpt),
		Content:        article.Content,
		ContentJSON:    parseJSON(article.ContentJson),
		CoverImageURL:  textValue(article.CoverImageUrl),
		Type:           article.Type,
		Status:         article.Status,
		SourceName:     textValue(article.SourceName),
		SourceURL:      textValue(article.SourceUrl),
		SEOTitle:       textValue(article.SeoTitle),
		SEODescription: textValue(article.SeoDescription),
		IsFeatured:     article.IsFeatured,
		AuthorID:       authorID.String(),
		Published:      timestamptzValue(article.PublishedAt),
	}, nil
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

func jsonBytes(value json.RawMessage) []byte {
	if len(value) == 0 {
		return nil
	}
	return []byte(value)
}

func chooseJSON(input json.RawMessage, fallback []byte) []byte {
	if len(input) == 0 {
		return fallback
	}
	return []byte(input)
}

func parseJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	var parsed any
	if err := json.Unmarshal(value, &parsed); err != nil {
		return nil
	}
	return parsed
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

func canManageArticle(actor Actor, article queries.Article) bool {
	if actor.Role == "admin" {
		return true
	}
	actorID, err := uuid.Parse(actor.UserID)
	if err != nil {
		return false
	}
	return sameUUID(article.AuthorID, actorID)
}

func canManageOpportunity(actor Actor, opportunity queries.Opportunity) bool {
	if actor.Role == "admin" {
		return true
	}
	actorID, err := uuid.Parse(actor.UserID)
	if err != nil {
		return false
	}
	return sameUUID(opportunity.PublishedBy, actorID)
}

func canManageUniversity(actor Actor, university queries.University) bool {
	if actor.Role == "admin" {
		return true
	}
	actorID, err := uuid.Parse(actor.UserID)
	if err != nil {
		return false
	}
	return sameUUID(university.CreatedBy, actorID)
}

func (s *Service) canManageCourse(ctx context.Context, actor Actor, course queries.Course) (bool, error) {
	if actor.Role == "admin" {
		return true, nil
	}
	university, err := s.repo.GetUniversityByID(ctx, course.UniversityID)
	if err != nil {
		return false, fmt.Errorf("cms: get university for course ownership: %w", err)
	}
	return canManageUniversity(actor, university), nil
}

func (s *Service) filterArticlesByActor(items []queries.Article, total int64, actor Actor, params PaginationParams) ([]queries.Article, int64, error) {
	if actor.Role == "admin" {
		return items, total, nil
	}
	filtered := make([]queries.Article, 0, len(items))
	for _, item := range items {
		if canManageArticle(actor, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func (s *Service) filterOpportunitiesByActor(items []queries.Opportunity, total int64, actor Actor, params PaginationParams) ([]queries.Opportunity, int64, error) {
	if actor.Role == "admin" {
		return items, total, nil
	}
	filtered := make([]queries.Opportunity, 0, len(items))
	for _, item := range items {
		if canManageOpportunity(actor, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func (s *Service) filterUniversitiesByActor(items []queries.University, total int64, actor Actor, params PaginationParams) ([]queries.University, int64, error) {
	if actor.Role == "admin" {
		return items, total, nil
	}
	filtered := make([]queries.University, 0, len(items))
	for _, item := range items {
		if canManageUniversity(actor, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func (s *Service) filterCoursesByActor(ctx context.Context, items []queries.Course, total int64, actor Actor, params PaginationParams) ([]queries.Course, int64, error) {
	if actor.Role == "admin" {
		return items, total, nil
	}
	filtered := make([]queries.Course, 0, len(items))
	for _, item := range items {
		ok, err := s.canManageCourse(ctx, actor, item)
		if err != nil {
			return nil, 0, err
		}
		if ok {
			filtered = append(filtered, item)
		}
	}
	return filtered, int64(len(filtered)), nil
}

func sameUUID(value pgtype.UUID, compare uuid.UUID) bool {
	if !value.Valid {
		return false
	}
	return uuid.UUID(value.Bytes) == compare
}

func filterArticles(items []queries.Article, filters ArticleListFilters) []queries.Article {
	filtered := make([]queries.Article, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	for _, item := range items {
		if filters.Type != "" && item.Type != filters.Type {
			continue
		}
		if filters.Status != "" && item.Status != filters.Status {
			continue
		}
		if filters.Featured != nil && item.IsFeatured != *filters.Featured {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.Content), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortArticles(items []queries.Article, sortKey string) {
	switch sortKey {
	case "title_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Title < items[j].Title })
	case "title_desc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Title > items[j].Title })
	case "created_at_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.Before(items[j].CreatedAt.Time) })
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.After(items[j].CreatedAt.Time) })
	}
}

func paginateArticles(items []queries.Article, page, perPage int) []queries.Article {
	return paginateAny(items, page, perPage)
}

func filterOpportunities(items []queries.Opportunity, filters OpportunityListFilters) []queries.Opportunity {
	filtered := make([]queries.Opportunity, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	for _, item := range items {
		if filters.Type != "" && item.Type != filters.Type {
			continue
		}
		if filters.Verified != nil && item.Verified != *filters.Verified {
			continue
		}
		if filters.Active != nil && item.IsActive != *filters.Active {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Title+" "+item.EntityName+" "+item.Description), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortOpportunities(items []queries.Opportunity, sortKey string) {
	switch sortKey {
	case "title_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Title < items[j].Title })
	case "title_desc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Title > items[j].Title })
	case "deadline_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Deadline.Time.Before(items[j].Deadline.Time) })
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.After(items[j].CreatedAt.Time) })
	}
}

func paginateOpportunities(items []queries.Opportunity, page, perPage int) []queries.Opportunity {
	return paginateAny(items, page, perPage)
}

func filterUniversities(items []queries.University, filters UniversityListFilters) []queries.University {
	filtered := make([]queries.University, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	for _, item := range items {
		if filters.Type != "" && item.Type != filters.Type {
			continue
		}
		if filters.Province != "" && item.Province != filters.Province {
			continue
		}
		if filters.Verified != nil && item.Verified != *filters.Verified {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Name+" "+textValue(item.Description)), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortUniversities(items []queries.University, sortKey string) {
	switch sortKey {
	case "name_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	case "name_desc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name > items[j].Name })
	case "created_at_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.Before(items[j].CreatedAt.Time) })
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.After(items[j].CreatedAt.Time) })
	}
}

func paginateUniversities(items []queries.University, page, perPage int) []queries.University {
	return paginateAny(items, page, perPage)
}

func filterCourses(items []queries.Course, filters CourseListFilters) []queries.Course {
	filtered := make([]queries.Course, 0, len(items))
	query := strings.ToLower(strings.TrimSpace(filters.Query))
	for _, item := range items {
		if filters.Area != "" && item.Area != filters.Area {
			continue
		}
		if filters.Level != "" && item.Level != filters.Level {
			continue
		}
		if filters.Regime != "" && item.Regime != filters.Regime {
			continue
		}
		if filters.UniversityID != "" && item.UniversityID.Valid && uuid.UUID(item.UniversityID.Bytes).String() != filters.UniversityID {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Name+" "+item.Area), query) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func sortCourses(items []queries.Course, sortKey string) {
	switch sortKey {
	case "name_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	case "name_desc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].Name > items[j].Name })
	case "created_at_asc":
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.Before(items[j].CreatedAt.Time) })
	default:
		sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.Time.After(items[j].CreatedAt.Time) })
	}
}

func paginateCourses(items []queries.Course, page, perPage int) []queries.Course {
	return paginateAny(items, page, perPage)
}

func paginateAny[T any](items []T, page, perPage int) []T {
	start := (page - 1) * perPage
	if start >= len(items) {
		return []T{}
	}
	end := start + perPage
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
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

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
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
