package articles

import (
	"context"
	"fmt"
	"math"
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

type Service struct{ repo Repository }

type PaginationParams struct{ Page, PerPage int }

type Filters struct {
	Query    string
	Type     string
	Featured *bool
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

type ArticleDetailResult struct {
	Data ArticleItem `json:"data"`
}

type ArticleItem struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt,omitempty"`
	Content        string `json:"content"`
	CoverImageURL  string `json:"cover_image_url,omitempty"`
	Type           string `json:"type"`
	SourceName     string `json:"source_name,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`
	SEOTitle       string `json:"seo_title,omitempty"`
	SEODescription string `json:"seo_description,omitempty"`
	IsFeatured     bool   `json:"is_featured"`
	PublishedAt    string `json:"published_at,omitempty"`
}

func NewService(repo Repository) *Service { return &Service{repo: repo} }

func (s *Service) ListArticles(ctx context.Context, params PaginationParams, filters Filters) (*ArticlesResult, error) {
	page, perPage := normalizePagination(params)
	items, err := s.repo.ListArticles(ctx, queries.ListArticlesParams{Limit: int32(perPage), Offset: int32((page - 1) * perPage)}, filters)
	if err != nil {
		return nil, fmt.Errorf("articles: list articles: %w", err)
	}
	total, err := s.repo.CountArticles(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("articles: count articles: %w", err)
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

func (s *Service) GetArticleBySlug(ctx context.Context, slug string) (*ArticleDetailResult, error) {
	item, err := s.repo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	mapped, err := mapArticle(item)
	if err != nil {
		return nil, err
	}
	return &ArticleDetailResult{Data: mapped}, nil
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

func mapArticle(item queries.Article) (ArticleItem, error) {
	id, err := uuidFromPg(item.ID)
	if err != nil {
		return ArticleItem{}, fmt.Errorf("articles: article id: %w", err)
	}
	return ArticleItem{
		ID:             id.String(),
		Slug:           item.Slug,
		Title:          item.Title,
		Excerpt:        textValue(item.Excerpt),
		Content:        item.Content,
		CoverImageURL:  textValue(item.CoverImageUrl),
		Type:           item.Type,
		SourceName:     textValue(item.SourceName),
		SourceURL:      textValue(item.SourceUrl),
		SEOTitle:       textValue(item.SeoTitle),
		SEODescription: textValue(item.SeoDescription),
		IsFeatured:     item.IsFeatured,
		PublishedAt:    timestamptzValue(item.PublishedAt),
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
	return value.Time.UTC().Format(time.RFC3339)
}
