package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db/queries"
)

type mockRepository struct {
	getArticleByIDFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	publishArticleFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	unpublishArticleFn      func(context.Context, pgtype.UUID) (queries.Article, error)
	archiveArticleFn        func(context.Context, pgtype.UUID) (queries.Article, error)
	getOpportunityByIDFn    func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	verifyOpportunityFn     func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	rejectOpportunityFn     func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	deactivateOpportunityFn func(context.Context, pgtype.UUID) (queries.Opportunity, error)
	listReportsFn           func(context.Context, queries.ListReportsParams, ReportListFilters) ([]queries.Report, error)
	countReportsFn          func(context.Context, ReportListFilters) (int64, error)
	getReportByIDFn         func(context.Context, pgtype.UUID) (queries.Report, error)
	updateReportStatusFn    func(context.Context, queries.UpdateReportStatusParams) (queries.Report, error)
	setUniversityVerifiedFn func(context.Context, queries.SetUniversityVerifiedParams) (queries.University, error)
	listUsersFn             func(context.Context, queries.AdminListUsersParams) ([]queries.AdminListUsersRow, error)
	getUserFn               func(context.Context, pgtype.UUID) (queries.AdminGetUserRow, error)
	updateUserRoleFn        func(context.Context, queries.AdminUpdateUserRoleParams) (queries.AdminUpdateUserRoleRow, error)
}

func (m *mockRepository) SetUniversityVerified(ctx context.Context, params queries.SetUniversityVerifiedParams) (queries.University, error) {
	return m.setUniversityVerifiedFn(ctx, params)
}

func (m *mockRepository) GetArticleByID(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.getArticleByIDFn(ctx, id)
}
func (m *mockRepository) PublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.publishArticleFn(ctx, id)
}
func (m *mockRepository) UnpublishArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.unpublishArticleFn(ctx, id)
}
func (m *mockRepository) ArchiveArticle(ctx context.Context, id pgtype.UUID) (queries.Article, error) {
	return m.archiveArticleFn(ctx, id)
}
func (m *mockRepository) GetOpportunityByID(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.getOpportunityByIDFn(ctx, id)
}
func (m *mockRepository) VerifyOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.verifyOpportunityFn(ctx, id)
}
func (m *mockRepository) RejectOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.rejectOpportunityFn(ctx, id)
}
func (m *mockRepository) DeactivateOpportunity(ctx context.Context, id pgtype.UUID) (queries.Opportunity, error) {
	return m.deactivateOpportunityFn(ctx, id)
}
func (m *mockRepository) ListReports(ctx context.Context, params queries.ListReportsParams, filters ReportListFilters) ([]queries.Report, error) {
	return m.listReportsFn(ctx, params, filters)
}
func (m *mockRepository) CountReports(ctx context.Context, filters ReportListFilters) (int64, error) {
	return m.countReportsFn(ctx, filters)
}
func (m *mockRepository) GetReportByID(ctx context.Context, id pgtype.UUID) (queries.Report, error) {
	return m.getReportByIDFn(ctx, id)
}
func (m *mockRepository) UpdateReportStatus(ctx context.Context, params queries.UpdateReportStatusParams) (queries.Report, error) {
	return m.updateReportStatusFn(ctx, params)
}
func (m *mockRepository) CountCMSArticles(context.Context) (int64, error)      { return 0, nil }
func (m *mockRepository) CountCMSOpportunities(context.Context) (int64, error) { return 0, nil }
func (m *mockRepository) CountPendingReports(context.Context) (int64, error)   { return 0, nil }
func (m *mockRepository) CountPendingMentorshipSessions(context.Context) (int64, error) {
	return 0, nil
}
func (m *mockRepository) ListRecentPendingReports(context.Context) ([]queries.Report, error) {
	return nil, nil
}
func (m *mockRepository) ListRecentUnverifiedOpportunities(context.Context) ([]queries.Opportunity, error) {
	return nil, nil
}
func (m *mockRepository) ContentCreatedPerDay(context.Context, int) ([]DailyContentCount, error) {
	return nil, nil
}
func (m *mockRepository) OpportunitiesByType(context.Context) ([]TypeCount, error) {
	return nil, nil
}
func (m *mockRepository) OpportunitiesByStatus(context.Context) ([]TypeCount, error) {
	return nil, nil
}

func TestHandlerPublishArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
		publishArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "published", PublishedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true}}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/publish", handler.PublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/publish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerVerifyOpportunityNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
		return queries.Opportunity{}, ErrNotFound
	}}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/verify", handler.VerifyOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+uuid.NewString()+"/verify", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerUniversityVerification(t *testing.T) {
	universityID := uuid.New()
	cases := []struct {
		action   string
		verified bool
	}{{"verify", true}, {"unverify", false}}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			var got queries.SetUniversityVerifiedParams
			handler := NewHandler(NewService(&mockRepository{
				setUniversityVerifiedFn: func(_ context.Context, params queries.SetUniversityVerifiedParams) (queries.University, error) {
					got = params
					return queries.University{ID: params.ID, Slug: "uem", Name: "UEM", Verified: params.Verified}, nil
				},
			}))
			app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
			app.Post("/v1/admin/universities/:id/verify", handler.VerifyUniversity)
			app.Post("/v1/admin/universities/:id/unverify", handler.UnverifyUniversity)
			req := httptest.NewRequest(http.MethodPost, "/v1/admin/universities/"+universityID.String()+"/"+tc.action, nil)
			res, err := app.Test(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if res.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", res.StatusCode)
			}
			if got.Verified != tc.verified || got.ID.Bytes != [16]byte(universityID) {
				t.Fatalf("unexpected params: %+v", got)
			}
		})
	}
}

func TestHandlerVerifyUniversityNotFound(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{
		setUniversityVerifiedFn: func(context.Context, queries.SetUniversityVerifiedParams) (queries.University, error) {
			return queries.University{}, ErrNotFound
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/universities/:id/verify", handler.VerifyUniversity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/universities/"+uuid.NewString()+"/verify", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlerPublishArticleValidatesID(t *testing.T) {
	handler := NewHandler(NewService(&mockRepository{}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/publish", handler.PublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/bad-id/publish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlerUnpublishArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "published"}, nil
		},
		unpublishArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/unpublish", handler.UnpublishArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/unpublish", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerArchiveArticle(t *testing.T) {
	articleID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getArticleByIDFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "draft"}, nil
		},
		archiveArticleFn: func(context.Context, pgtype.UUID) (queries.Article, error) {
			return queries.Article{ID: pgtype.UUID{Bytes: [16]byte(articleID), Valid: true}, Title: "Artigo", Slug: "artigo", Status: "archived"}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/articles/:id/archive", handler.ArchiveArticle)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/articles/"+articleID.String()+"/archive", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerRejectOpportunity(t *testing.T) {
	opportunityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: true}, nil
		},
		rejectOpportunityFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: false, IsActive: false}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/reject", handler.RejectOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+opportunityID.String()+"/reject", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandlerDeactivateOpportunity(t *testing.T) {
	opportunityID := uuid.New()
	handler := NewHandler(NewService(&mockRepository{
		getOpportunityByIDFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: true}, nil
		},
		deactivateOpportunityFn: func(context.Context, pgtype.UUID) (queries.Opportunity, error) {
			return queries.Opportunity{ID: pgtype.UUID{Bytes: [16]byte(opportunityID), Valid: true}, Title: "Opp", Slug: "opp", Verified: true, IsActive: false}, nil
		},
	}))
	app := fiber.New(fiber.Config{ErrorHandler: apierror.Handler})
	app.Post("/v1/admin/opportunities/:id/deactivate", handler.DeactivateOpportunity)
	req := httptest.NewRequest(http.MethodPost, "/v1/admin/opportunities/"+opportunityID.String()+"/deactivate", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func (m *mockRepository) ListUsers(ctx context.Context, params queries.AdminListUsersParams) ([]queries.AdminListUsersRow, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx, params)
	}
	return nil, nil
}
func (m *mockRepository) CountUsers(context.Context, queries.AdminCountUsersParams) (int64, error) {
	return 0, nil
}
func (m *mockRepository) CountUsersByRole(context.Context) ([]queries.AdminCountUsersByRoleRow, error) {
	return []queries.AdminCountUsersByRoleRow{{Role: "admin", Total: 1}}, nil
}
func (m *mockRepository) GetUser(ctx context.Context, id pgtype.UUID) (queries.AdminGetUserRow, error) {
	if m.getUserFn != nil {
		return m.getUserFn(ctx, id)
	}
	return queries.AdminGetUserRow{}, ErrNotFound
}
func (m *mockRepository) UpdateUserRole(ctx context.Context, params queries.AdminUpdateUserRoleParams) (queries.AdminUpdateUserRoleRow, error) {
	if m.updateUserRoleFn != nil {
		return m.updateUserRoleFn(ctx, params)
	}
	return queries.AdminUpdateUserRoleRow{}, nil
}
