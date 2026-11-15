package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	appaccount "github.com/rnrnshn/oportunidades-api/internal/account"
	appadmin "github.com/rnrnshn/oportunidades-api/internal/admin"
	apparticles "github.com/rnrnshn/oportunidades-api/internal/articles"
	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	appcatalog "github.com/rnrnshn/oportunidades-api/internal/catalog"
	appcms "github.com/rnrnshn/oportunidades-api/internal/cms"
	appmentorship "github.com/rnrnshn/oportunidades-api/internal/mentorship"
	appopportunities "github.com/rnrnshn/oportunidades-api/internal/opportunities"
	appreports "github.com/rnrnshn/oportunidades-api/internal/reports"
	appsaved "github.com/rnrnshn/oportunidades-api/internal/saved"
	appuploads "github.com/rnrnshn/oportunidades-api/internal/uploads"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/db"
	appemail "github.com/rnrnshn/oportunidades-api/pkg/email"
	appmiddleware "github.com/rnrnshn/oportunidades-api/pkg/middleware"
	appstorage "github.com/rnrnshn/oportunidades-api/pkg/storage"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(getEnv("PORT", "8080")); err != nil {
			fmt.Fprintln(os.Stderr, "unhealthy:", err)
			os.Exit(1)
		}
		return
	}

	log := newLogger()

	if isProduction(getEnv("ENV", "development")) {
		if err := validateProductionConfig(os.Getenv); err != nil {
			log.Fatal().Err(err).Msg("refusing to start")
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	databaseURL := getEnv("DATABASE_URL", "postgresql://postgres:postgres@localhost:5432/oportunidades?sslmode=disable")
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	appConfig := fiber.Config{
		ErrorHandler: apierror.Handler,
		AppName:      "oportunidades-api",
		BodyLimit:    2 * 1024 * 1024,
	}
	if isProduction(getEnv("ENV", "development")) {
		// Cloudflare overwrites this header and the API is only reachable through the tunnel and gateway.
		appConfig.ProxyHeader = "CF-Connecting-IP"
	}
	app := fiber.New(appConfig)

	app.Use(appmiddleware.Logger(log))
	app.Use(appmiddleware.Recover())
	app.Use(appmiddleware.SecurityHeaders())
	app.Use(appmiddleware.CORS(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")))

	registerRoutes(app, pool, log)

	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			log.Error().Err(err).Msg("database health check failed")
			return apierror.Internal()
		}

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"status":  "ok",
				"service": "oportunidades-api",
			},
		})
	})

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("failed to shut down server")
		}
	}()

	port := getEnv("PORT", "8080")
	if err := app.Listen(":" + port); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}

func newLogger() zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if strings.EqualFold(getEnv("ENV", "development"), "development") {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	return logger
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func registerRoutes(app *fiber.App, pool *pgxpool.Pool, log zerolog.Logger) {
	authRepository := appauth.NewPostgresRepository(pool)
	var emailSender appemail.Sender
	if resendAPIKey := getEnv("RESEND_API_KEY", ""); resendAPIKey != "" {
		emailSender = appemail.NewResendSender(resendAPIKey, getEnv("EMAIL_FROM", ""))
	}
	authService := appauth.NewService(authRepository, appauth.Config{
		JWTSecret:             getEnv("JWT_SECRET", "change-me"),
		JWTExpiry:             time.Duration(getEnvAsInt("JWT_EXPIRY_MINUTES", 15)) * time.Minute,
		RefreshTokenExpiry:    time.Duration(getEnvAsInt("REFRESH_TOKEN_EXPIRY_DAYS", 30)) * 24 * time.Hour,
		AuthActionTokenExpiry: time.Duration(getEnvAsInt("AUTH_ACTION_TOKEN_EXPIRY_HOURS", 24)) * time.Hour,
		RefreshCookieName:     "refresh_token",
		RefreshCookieSecure:   strings.EqualFold(getEnv("ENV", "development"), "production"),
		ExposeDebugTokens:     !strings.EqualFold(getEnv("ENV", "development"), "production"),
		EmailSender:           emailSender,
		AppPublicURL:          getEnv("APP_PUBLIC_URL", "http://localhost:3000"),
		Logger:                &log,
	})
	authHandler := appauth.NewHandler(authService)
	accountRepository := appaccount.NewPostgresRepository(pool)
	accountService := appaccount.NewService(accountRepository)
	accountHandler := appaccount.NewHandler(accountService)
	adminRepository := appadmin.NewPostgresRepository(pool)
	adminService := appadmin.NewService(adminRepository, appadmin.WithInvites(emailSender, getEnv("DASHBOARD_URL", "http://localhost:3002")))
	adminHandler := appadmin.NewHandler(adminService)
	articlesRepository := apparticles.NewPostgresRepository(pool)
	articlesService := apparticles.NewService(articlesRepository)
	articlesHandler := apparticles.NewHandler(articlesService)
	catalogRepository := appcatalog.NewPostgresRepository(pool)
	catalogService := appcatalog.NewService(catalogRepository)
	catalogHandler := appcatalog.NewHandler(catalogService)
	cmsRepository := appcms.NewPostgresRepository(pool)
	cmsService := appcms.NewService(cmsRepository)
	cmsHandler := appcms.NewHandler(cmsService)
	mentorshipRepository := appmentorship.NewPostgresRepository(pool)
	mentorshipService := appmentorship.NewService(mentorshipRepository)
	mentorshipHandler := appmentorship.NewHandler(mentorshipService)
	opportunitiesRepository := appopportunities.NewPostgresRepository(pool)
	opportunitiesService := appopportunities.NewService(opportunitiesRepository)
	opportunitiesHandler := appopportunities.NewHandler(opportunitiesService)
	reportsRepository := appreports.NewPostgresRepository(pool)
	reportsService := appreports.NewService(reportsRepository)
	reportsHandler := appreports.NewHandler(reportsService)
	savedHandler := appsaved.NewHandler(appsaved.NewService(appsaved.NewPostgresRepository(pool)))
	storageClient := appstorage.NewSupabaseClient(getEnv("SUPABASE_URL", ""), getEnv("SUPABASE_STORAGE_BUCKET", ""), getEnv("SUPABASE_SERVICE_ROLE_KEY", ""))
	uploadsService := appuploads.NewService(storageClient)
	uploadsHandler := appuploads.NewHandler(uploadsService)

	v1 := app.Group("/v1")
	authGroup := v1.Group("/auth")
	authLimit := appmiddleware.RateLimit(10, time.Minute)
	emailLimit := appmiddleware.RateLimit(5, 15*time.Minute)
	authGroup.Post("/register", authLimit, authHandler.Register)
	authGroup.Post("/login", authLimit, authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Post("/forgot-password", emailLimit, authHandler.ForgotPassword)
	authGroup.Post("/reset-password", authLimit, authHandler.ResetPassword)
	authGroup.Post("/verify-email", authLimit, authHandler.VerifyEmail)
	authGroup.Post("/send-verification", emailLimit, appauth.RequireAuth(authService), authHandler.SendVerification)

	accountGroup := v1.Group("/account", appauth.RequireAuth(authService))
	accountGroup.Get("/me", accountHandler.GetMe)
	accountGroup.Patch("/me", accountHandler.UpdateMe)
	accountGroup.Post("/password", accountHandler.ChangePassword)
	accountGroup.Post("/logout-all", authHandler.LogoutAll)
	accountGroup.Post("/deactivate", authHandler.DeactivateAccount)
	accountGroup.Get("/saved", savedHandler.List)
	accountGroup.Get("/saved/ids", savedHandler.IDs)
	accountGroup.Put("/saved/:opportunityId", savedHandler.Save)
	accountGroup.Delete("/saved/:opportunityId", savedHandler.Unsave)

	cmsGroup := v1.Group("/cms", appauth.RequireRole(authService, "cms_partner", "admin"))
	cmsGroup.Get("/articles", cmsHandler.ListArticles)
	cmsGroup.Get("/articles/:id", cmsHandler.GetArticle)
	cmsGroup.Post("/articles", cmsHandler.CreateArticle)
	cmsGroup.Patch("/articles/:id", cmsHandler.UpdateArticle)
	cmsGroup.Get("/universities", cmsHandler.ListUniversities)
	cmsGroup.Get("/universities/:id", cmsHandler.GetUniversity)
	cmsGroup.Post("/universities", cmsHandler.CreateUniversity)
	cmsGroup.Patch("/universities/:id", cmsHandler.UpdateUniversity)
	cmsGroup.Get("/courses", cmsHandler.ListCourses)
	cmsGroup.Get("/courses/:id", cmsHandler.GetCourse)
	cmsGroup.Post("/courses", cmsHandler.CreateCourse)
	cmsGroup.Patch("/courses/:id", cmsHandler.UpdateCourse)
	cmsGroup.Get("/opportunities", cmsHandler.ListOpportunities)
	cmsGroup.Get("/opportunities/:id", cmsHandler.GetOpportunity)
	cmsGroup.Post("/opportunities", cmsHandler.CreateOpportunity)
	cmsGroup.Patch("/opportunities/:id", cmsHandler.UpdateOpportunity)

	adminGroup := v1.Group("/admin", appauth.RequireRole(authService, "admin"))
	adminGroup.Get("/dashboard", adminHandler.Dashboard)
	adminGroup.Get("/dashboard/analytics", adminHandler.Analytics)
	adminGroup.Post("/articles/:id/publish", adminHandler.PublishArticle)
	adminGroup.Post("/articles/:id/unpublish", adminHandler.UnpublishArticle)
	adminGroup.Post("/articles/:id/archive", adminHandler.ArchiveArticle)
	adminGroup.Post("/opportunities/:id/verify", adminHandler.VerifyOpportunity)
	adminGroup.Post("/opportunities/:id/reject", adminHandler.RejectOpportunity)
	adminGroup.Post("/opportunities/:id/deactivate", adminHandler.DeactivateOpportunity)
	adminGroup.Post("/universities/:id/verify", adminHandler.VerifyUniversity)
	adminGroup.Post("/universities/:id/unverify", adminHandler.UnverifyUniversity)
	adminGroup.Get("/reports", adminHandler.ListReports)
	adminGroup.Get("/reports/:id", adminHandler.GetReport)
	adminGroup.Patch("/reports/:id", adminHandler.UpdateReportStatus)
	adminGroup.Get("/users", adminHandler.ListUsers)
	adminGroup.Get("/users/:id", adminHandler.GetUser)
	adminGroup.Patch("/users/:id/role", adminHandler.UpdateUserRole)
	adminGroup.Post("/users/:id/suspend", adminHandler.SuspendUser)
	adminGroup.Post("/users/:id/unsuspend", adminHandler.UnsuspendUser)
	adminGroup.Get("/invites", adminHandler.ListInvites)
	inviteLimit := appmiddleware.RateLimit(30, time.Hour)
	adminGroup.Post("/invites", inviteLimit, adminHandler.CreateInvite)
	adminGroup.Post("/invites/:id/resend", inviteLimit, adminHandler.ResendInvite)
	adminGroup.Delete("/invites/:id", adminHandler.RevokeInvite)
	invitesGroup := v1.Group("/invites")
	invitesGroup.Get("/:token", authLimit, adminHandler.PublicInvite)
	invitesGroup.Post("/:token/accept", authLimit, adminHandler.AcceptInvite)

	articlesGroup := v1.Group("/articles")
	articlesGroup.Get("", articlesHandler.ListArticles)
	articlesGroup.Get("/:slug", articlesHandler.GetArticleBySlug)

	catalogGroup := v1.Group("/catalog")
	catalogGroup.Get("/universities", catalogHandler.ListUniversities)
	catalogGroup.Get("/universities/:slug", catalogHandler.GetUniversityBySlug)
	catalogGroup.Get("/courses", catalogHandler.ListCourses)
	catalogGroup.Get("/courses/:slug", catalogHandler.GetCourseBySlug)

	mentorshipGroup := v1.Group("/mentorship")
	mentorshipGroup.Get("/mentors", mentorshipHandler.ListMentors)
	mentorshipGroup.Get("/mentors/:id", mentorshipHandler.GetMentorByID)
	mentorshipGroup.Post("/sessions", appauth.RequireAuth(authService), mentorshipHandler.CreateSessionRequest)
	mentorshipGroup.Get("/sessions", appauth.RequireAuth(authService), mentorshipHandler.ListSessions)
	mentorshipGroup.Get("/sessions/:id", appauth.RequireAuth(authService), mentorshipHandler.GetSession)
	mentorshipGroup.Patch("/sessions/:id", appauth.RequireAuth(authService), mentorshipHandler.UpdateSessionStatus)

	opportunitiesGroup := v1.Group("/opportunities")
	opportunitiesGroup.Get("", opportunitiesHandler.ListOpportunities)
	opportunitiesGroup.Get("/:slug", opportunitiesHandler.GetOpportunityBySlug)

	reportsGroup := v1.Group("/reports", appauth.RequireAuth(authService))
	reportsGroup.Post("", appmiddleware.RateLimit(20, time.Hour), reportsHandler.Create)

	uploadsGroup := v1.Group("/uploads", appauth.RequireAuth(authService))
	uploadsGroup.Post("/presign", uploadsHandler.Presign)
	uploadsGroup.Post("/confirm", uploadsHandler.Confirm)
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
