package cms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	createArticleFn         func(context.Context, queries.CreateArticleParams) (queries.Article, error)
	createOpportunityFn     func(context.Context, queries.CreateOpportunityParams) (queries.Opportunity, error)
	listCMSArticlesFn       func(context.Context, queries.ListCMSArticlesParams, Actor, ArticleListFilters) ([]queries.Article, error)
	countCMSArticlesFn      func(context.Context, Actor, ArticleListFilters) (int64, error)
	getArticleByIDFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	updateArticleFn         func(context.Context, queries.UpdateArticleParams) (queries.Article, error)
	createUniversityFn      func(context.Context, queries.CreateUniversityParams) (queries.University, error)
	listCMSUniversitiesFn   func(context.Context, queries.ListCMSUniversitiesParams, Actor, UniversityListFilters) ([]queries.University, error)
	countCMSUniversitiesFn  func(context.Context, Actor, UniversityListFilters) (int64, error)
	getUniversityByIDFn     func(context.Context, pgtype.UUID) (queries.University, error)
	updateUniversityFn      func(context.Context, queries.UpdateUniversityParams) (queries.University, error)
	createCourseFn          func(context.Context, queries.CreateCourseParams) (queries.Course, error)
	listCMSCoursesFn        func(context.Context, queries.ListCMSCoursesParams, Actor, CourseListFilters) ([]queries.Course, error)
	countCMSCoursesFn       func(context.Context, Actor, CourseListFilters) (int64, error)
	getCourseByIDFn         func(context.Context, pgtype.UUID) (queries.Course, error)
	updateCourseFn          func(context.Context, queries.UpdateCourseParams) (queries.Course, error)
	listCMSOpportunitiesFn  func(context.Context, queries.ListCMSOpportunitiesParams, Actor, OpportunityListFilters) ([]queries.Opportunity, error)
	countCMSOpportunitiesFn func(context.Context, Actor, OpportunityListFilters) (int64, error)
	getOpportunityByIDFn    func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	updateOpportunityFn     func(context.Context, queries.UpdateOpportunityParams) (queries.Opportunity, error)
}

func (m *mockRepository) CreateArticle(ctx context.Context, params queries.CreateArticleParams) (queries.Article, error) {
	return m.createArticleFn(ctx, params)
}
func (m *mockRepository) CreateOpportunity(ctx context.Context, params queries.CreateOpportunityParams) (queries.Opportunity, error) {
	return m.createOpportunityFn(ctx, params)
}
func (m *mockRepository) ListCMSArticles(ctx context.Context, params queries.ListCMSArticlesParams, actor Actor, filters ArticleListFilters) ([]queries.Article, error) {
	return m.listCMSArticlesFn(ctx, params, actor, filters)
}
func (m *mockRepository) CountCMSArticles(ctx context.Context, actor Actor, filters ArticleListFilters) (int64, error) {
	return m.countCMSArticlesFn(ctx, actor, filters)
}
func (m *mockRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.getArticleByIDFn(ctx, id)
}
func (m *mockRepository) UpdateArticle(ctx context.Context, params queries.UpdateArticleParams) (queries.Article, error) {
	return m.updateArticleFn(ctx, params)
}
func (m *mockRepository) CreateUniversity(ctx context.Context, params queries.CreateUniversityParams) (queries.University, error) {
	return m.createUniversityFn(ctx, params)
}
func (m *mockRepository) ListCMSUniversities(ctx context.Context, params queries.ListCMSUniversitiesParams, actor Actor, filters UniversityListFilters) ([]queries.University, error) {
	return m.listCMSUniversitiesFn(ctx, params, actor, filters)
}
func (m *mockRepository) CountCMSUniversities(ctx context.Context, actor Actor, filters UniversityListFilters) (int64, error) {
	return m.countCMSUniversitiesFn(ctx, actor, filters)
}
func (m *mockRepository) GetUniversityByID(ctx context.Context, id pgtype.UUID) (queries.University, error) {
	return m.getUniversityByIDFn(ctx, id)
}
func (m *mockRepository) UpdateUniversity(ctx context.Context, params queries.UpdateUniversityParams) (queries.University, error) {
	return m.updateUniversityFn(ctx, params)
}
func (m *mockRepository) CreateCourse(ctx context.Context, params queries.CreateCourseParams) (queries.Course, error) {
	return m.createCourseFn(ctx, params)
}
func (m *mockRepository) ListCMSCourses(ctx context.Context, params queries.ListCMSCoursesParams, actor Actor, filters CourseListFilters) ([]queries.Course, error) {
	return m.listCMSCoursesFn(ctx, params, actor, filters)
}
func (m *mockRepository) CountCMSCourses(ctx context.Context, actor Actor, filters CourseListFilters) (int64, error) {
	return m.countCMSCoursesFn(ctx, actor, filters)
}
func (m *mockRepository) GetCourseByID(ctx context.Context, id pgtype.UUID) (queries.Course, error) {
	return m.getCourseByIDFn(ctx, id)
}
func (m *mockRepository) UpdateCourse(ctx context.Context, params queries.UpdateCourseParams) (queries.Course, error) {
	return m.updateCourseFn(ctx, params)
}
func (m *mockRepository) ListCMSOpportunities(ctx context.Context, params queries.ListCMSOpportunitiesParams, actor Actor, filters OpportunityListFilters) ([]queries.Opportunity, error) {
	return m.listCMSOpportunitiesFn(ctx, params, actor, filters)
}
func (m *mockRepository) CountCMSOpportunities(ctx context.Context, actor Actor, filters OpportunityListFilters) (int64, error) {
	return m.countCMSOpportunitiesFn(ctx, actor, filters)
}
func (m *mockRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.getOpportunityByIDFn(ctx, id)
}
func (m *mockRepository) UpdateOpportunity(ctx context.Context, params queries.UpdateOpportunityParams) (queries.Opportunity, error) {
	return m.updateOpportunityFn(ctx, params)
}

func TestHandlerCreateArticle(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{createArticleFn: func(context.Context, queries.CreateArticleParams) (queries.Article, error) {
		return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Novo Artigo", Type: "guide", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/articles", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateArticle(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/articles", strings.NewReader(`{"title":"Novo Artigo","content":"Conteudo","type":"guide"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestHandlerCreateOpportunityValidation(t *testing.T) {
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/opportunities", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateOpportunity(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/opportunities", strings.NewReader(`{"title":""}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateArticle(t *testing.T) {
	articleID := uuid.New()
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Novo Artigo", Content: "Conteudo", Type: "guide", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
		},
		updateArticleFn: func(context.Context, queries.UpdateArticleParams) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "novo-artigo", Title: "Artigo Editado", Content: "Conteudo editado", Type: "news", Status: "draft", AuthorID: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/cms/articles/:id", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.UpdateArticle(c)
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/cms/articles/"+articleID.String(), strings.NewReader(`{"title":"Artigo Editado","content":"Conteudo editado","type":"news"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerUpdateOpportunityValidatesID(t *testing.T) {
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Patch("/v1/cms/opportunities/:id", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.UpdateOpportunity(c)
	})
	req := httptest.NewRequest(http.MethodPatch, "/v1/cms/opportunities/bad-id", strings.NewReader(`{"title":"Teste"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerCreateUniversity(t *testing.T) {
	universityID := uuid.New()
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{createUniversityFn: func(context.Context, queries.CreateUniversityParams) (queries.University, error) {
		return queries.University{ID: pgtype.UUID{Bytes: [16]byte(universityID), Valid: true}, Slug: "uni", Name: "Universidade Teste", Type: "publica", Province: "Maputo", CreatedBy: pgtype.UUID{Bytes: [16]byte(userID), Valid: true}}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/universities", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateUniversity(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/universities", strings.NewReader(`{"name":"Universidade Teste","type":"publica","province":"Maputo"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", res.StatusCode)
	}
}

func TestHandlerCreateCourseValidatesUniversityID(t *testing.T) {
	userID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/cms/courses", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: userID.String(), Role: "cms_partner"})
		return handler.CreateCourse(c)
	})
	req := httptest.NewRequest(http.MethodPost, "/v1/cms/courses", strings.NewReader(`{"university_id":"bad-id","name":"Curso","area":"Tecnologia","level":"licenciatura","regime":"presencial"}`))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerGetArticleForbiddenForOtherCmsPartner(t *testing.T) {
	articleID := uuid.New()
	ownerID := uuid.New()
	otherID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
		return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Slug: "a", Title: "A", Content: "c", Type: "guide", AuthorID: pgtype.UUID{Bytes: [16]byte(ownerID), Valid: true}}, nil
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Get("/v1/cms/articles/:id", func(c *fiber.Ctx) error {
		c.Locals("auth_user", appauth.AuthenticatedUser{ID: otherID.String(), Role: "cms_partner"})
		return handler.GetArticle(c)
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/cms/articles/"+articleID.String(), nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.StatusCode)
	}
}
