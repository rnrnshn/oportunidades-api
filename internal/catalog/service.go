package catalog

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

type UniversityFilters struct {
	Query    string
	Province string
	Type     string
	Verified *bool
}

type CourseFilters struct {
	Query        string
	Area         string
	Level        string
	Regime       string
	Province     string
	UniversityID string
}

type UniversitiesResult struct {
	Data []UniversityItem `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}

type CoursesResult struct {
	Data []CourseItem   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type UniversityDetailResult struct {
	Data UniversityItem `json:"data"`
}

type CourseDetailResult struct {
	Data CourseItem `json:"data"`
}

type UniversityItem struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Province    string `json:"province"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	Website     string `json:"website,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Verified    bool   `json:"verified"`
}

type CourseItem struct {
	ID                string `json:"id"`
	Slug              string `json:"slug"`
	UniversityID      string `json:"university_id"`
	Name              string `json:"name"`
	Area              string `json:"area"`
	Level             string `json:"level"`
	Regime            string `json:"regime"`
	DurationYears     *int32 `json:"duration_years,omitempty"`
	AnnualFee         string `json:"annual_fee,omitempty"`
	EntryRequirements string `json:"entry_requirements,omitempty"`
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUniversities(ctx context.Context, params PaginationParams, filters UniversityFilters) (*UniversitiesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListUniversities(ctx, queries.ListUniversitiesParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	}, filters)
	if err != nil {
		return nil, fmt.Errorf("catalog: list universities: %w", err)
	}

	total, err := s.repo.CountUniversities(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("catalog: count universities: %w", err)
	}

	data := make([]UniversityItem, 0, len(items))
	for _, item := range items {
		mappedItem, err := mapUniversity(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mappedItem)
	}

	return &UniversitiesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetUniversityBySlug(ctx context.Context, slug string) (*UniversityDetailResult, error) {
	item, err := s.repo.GetUniversityBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	mappedItem, err := mapUniversity(item)
	if err != nil {
		return nil, err
	}

	return &UniversityDetailResult{Data: mappedItem}, nil
}

func (s *Service) ListCourses(ctx context.Context, params PaginationParams, filters CourseFilters) (*CoursesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListCourses(ctx, queries.ListCoursesParams{
		Limit:  int32(perPage),
		Offset: int32((page - 1) * perPage),
	}, filters)
	if err != nil {
		return nil, fmt.Errorf("catalog: list courses: %w", err)
	}

	total, err := s.repo.CountCourses(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("catalog: count courses: %w", err)
	}

	data := make([]CourseItem, 0, len(items))
	for _, item := range items {
		mappedItem, err := mapCourse(item)
		if err != nil {
			return nil, err
		}
		data = append(data, mappedItem)
	}

	return &CoursesResult{Data: data, Meta: buildMeta(total, page, perPage)}, nil
}

func (s *Service) GetCourseBySlug(ctx context.Context, slug string) (*CourseDetailResult, error) {
	item, err := s.repo.GetCourseBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	mappedItem, err := mapCourse(item)
	if err != nil {
		return nil, err
	}

	return &CourseDetailResult{Data: mappedItem}, nil
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

func mapUniversity(item queries.University) (UniversityItem, error) {
	id, err := uuidFromPg(item.ID)
	if err != nil {
		return UniversityItem{}, fmt.Errorf("catalog: university id: %w", err)
	}

	return UniversityItem{
		ID:          id.String(),
		Slug:        item.Slug,
		Name:        item.Name,
		Type:        item.Type,
		Province:    item.Province,
		Description: textValue(item.Description),
		LogoURL:     textValue(item.LogoUrl),
		Website:     textValue(item.Website),
		Email:       textValue(item.Email),
		Phone:       textValue(item.Phone),
		Verified:    item.Verified,
	}, nil
}

func mapCourse(item queries.Course) (CourseItem, error) {
	id, err := uuidFromPg(item.ID)
	if err != nil {
		return CourseItem{}, fmt.Errorf("catalog: course id: %w", err)
	}

	universityID, err := uuidFromPg(item.UniversityID)
	if err != nil {
		return CourseItem{}, fmt.Errorf("catalog: course university id: %w", err)
	}

	return CourseItem{
		ID:                id.String(),
		Slug:              item.Slug,
		UniversityID:      universityID.String(),
		Name:              item.Name,
		Area:              item.Area,
		Level:             item.Level,
		Regime:            item.Regime,
		DurationYears:     int32Pointer(item.DurationYears),
		AnnualFee:         numericValue(item.AnnualFee),
		EntryRequirements: textValue(item.EntryRequirements),
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

func int32Pointer(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}

	result := value.Int32
	return &result
}

func numericValue(value pgtype.Numeric) string {
	if !value.Valid {
		return ""
	}

	encodedValue, err := value.Value()
	if err != nil || encodedValue == nil {
		return ""
	}

	formattedValue, ok := encodedValue.(string)
	if !ok {
		return ""
	}

	return formattedValue
}
