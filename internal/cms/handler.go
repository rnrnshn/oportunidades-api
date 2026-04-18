package cms

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct{ service *Service }

type createArticleRequest struct {
	Title          string `json:"title"`
	Excerpt        string `json:"excerpt"`
	Content        string `json:"content"`
	ContentJSON    json.RawMessage `json:"content_json"`
	CoverImageURL  string `json:"cover_image_url"`
	Type           string `json:"type"`
	SourceName     string `json:"source_name"`
	SourceURL      string `json:"source_url"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	IsFeatured     *bool  `json:"is_featured"`
}

type createOpportunityRequest struct {
	Title        string `json:"title"`
	Type         string `json:"type"`
	EntityName   string `json:"entity_name"`
	Description  string `json:"description"`
	Requirements string `json:"requirements"`
	Deadline     string `json:"deadline"`
	ApplyURL     string `json:"apply_url"`
	Country      string `json:"country"`
	Language     string `json:"language"`
	Area         string `json:"area"`
}

type createUniversityRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Province    string `json:"province"`
	Description string `json:"description"`
	LogoURL     string `json:"logo_url"`
	Website     string `json:"website"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

type createCourseRequest struct {
	UniversityID      string `json:"university_id"`
	Name              string `json:"name"`
	Area              string `json:"area"`
	Level             string `json:"level"`
	Regime            string `json:"regime"`
	DurationYears     int32  `json:"duration_years"`
	AnnualFee         string `json:"annual_fee"`
	EntryRequirements string `json:"entry_requirements"`
}

type updateArticleRequest = createArticleRequest
type updateOpportunityRequest = createOpportunityRequest
type updateUniversityRequest = createUniversityRequest
type updateCourseRequest = createCourseRequest

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListArticles(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("type", c.Query("type"), []string{"editorial", "news", "guide"}, "type deve ser editorial, news ou guide.")
	validationErrors.Enum("status", c.Query("status"), []string{"draft", "in_review", "published", "archived"}, "status inválido.")
	validationErrors.Bool("featured", c.Query("featured"), "featured deve ser true ou false.")
	validationErrors.Enum("sort", c.Query("sort"), []string{"title_asc", "title_desc", "created_at_asc"}, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListArticles(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, PaginationParams{Page: queryInt(c, "page", defaultPage), PerPage: queryInt(c, "per_page", defaultPerPage)}, ArticleListFilters{Query: strings.TrimSpace(c.Query("q")), Type: strings.TrimSpace(c.Query("type")), Status: strings.TrimSpace(c.Query("status")), Featured: queryBool(c, "featured"), Sort: strings.TrimSpace(c.Query("sort"))})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetArticle(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.GetArticle(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) CreateArticle(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request createArticleRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("title", request.Title, "Título é obrigatório.")
	validationErrors.Required("content", request.Content, "Conteúdo é obrigatório.")
	validationErrors.Required("type", request.Type, "Tipo é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.CreateArticle(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateArticleInput{
		AuthorID: currentUser.ID,
		Title:    strings.TrimSpace(request.Title), Excerpt: strings.TrimSpace(request.Excerpt), Content: strings.TrimSpace(request.Content), ContentJSON: request.ContentJSON, CoverImageURL: strings.TrimSpace(request.CoverImageURL),
		Type: strings.TrimSpace(request.Type), SourceName: strings.TrimSpace(request.SourceName), SourceURL: strings.TrimSpace(request.SourceURL), SEOTitle: strings.TrimSpace(request.SEOTitle), SEODescription: strings.TrimSpace(request.SEODescription), IsFeatured: request.IsFeatured,
	})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handler) CreateOpportunity(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request createOpportunityRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("title", request.Title, "Título é obrigatório.")
	validationErrors.Required("type", request.Type, "Tipo é obrigatório.")
	validationErrors.Required("entity_name", request.EntityName, "Entidade é obrigatória.")
	validationErrors.Required("description", request.Description, "Descrição é obrigatória.")
	validationErrors.Required("country", request.Country, "País é obrigatório.")
	validationErrors.RFC3339("deadline", request.Deadline, "deadline deve estar em formato RFC3339.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.CreateOpportunity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateOpportunityInput{
		PublishedBy: currentUser.ID,
		Title:       strings.TrimSpace(request.Title), Type: strings.TrimSpace(request.Type), EntityName: strings.TrimSpace(request.EntityName), Description: strings.TrimSpace(request.Description), Requirements: strings.TrimSpace(request.Requirements), Deadline: strings.TrimSpace(request.Deadline), ApplyURL: strings.TrimSpace(request.ApplyURL), Country: strings.TrimSpace(request.Country), Language: strings.TrimSpace(request.Language), Area: strings.TrimSpace(request.Area),
	})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handler) ListUniversities(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("type", c.Query("type"), []string{"publica", "privada", "instituto", "academia"}, "type deve ser publica, privada, instituto ou academia.")
	validationErrors.Bool("verified", c.Query("verified"), "verified deve ser true ou false.")
	validationErrors.Enum("sort", c.Query("sort"), []string{"name_asc", "name_desc", "created_at_asc"}, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListUniversities(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, PaginationParams{Page: queryInt(c, "page", defaultPage), PerPage: queryInt(c, "per_page", defaultPerPage)}, UniversityListFilters{Query: strings.TrimSpace(c.Query("q")), Type: strings.TrimSpace(c.Query("type")), Province: strings.TrimSpace(c.Query("province")), Verified: queryBool(c, "verified"), Sort: strings.TrimSpace(c.Query("sort"))})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetUniversity(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.GetUniversity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) CreateUniversity(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request createUniversityRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("name", request.Name, "Nome é obrigatório.")
	validationErrors.Required("type", request.Type, "Tipo é obrigatório.")
	validationErrors.Required("province", request.Province, "Província é obrigatória.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.CreateUniversity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateUniversityInput{CreatedBy: currentUser.ID, Name: strings.TrimSpace(request.Name), Type: strings.TrimSpace(request.Type), Province: strings.TrimSpace(request.Province), Description: strings.TrimSpace(request.Description), LogoURL: strings.TrimSpace(request.LogoURL), Website: strings.TrimSpace(request.Website), Email: strings.TrimSpace(request.Email), Phone: strings.TrimSpace(request.Phone)})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handler) UpdateUniversity(c *fiber.Ctx) error {
	var request updateUniversityRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.UpdateUniversity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateUniversityInput{ID: strings.TrimSpace(c.Params("id")), Name: strings.TrimSpace(request.Name), Type: strings.TrimSpace(request.Type), Province: strings.TrimSpace(request.Province), Description: strings.TrimSpace(request.Description), LogoURL: strings.TrimSpace(request.LogoURL), Website: strings.TrimSpace(request.Website), Email: strings.TrimSpace(request.Email), Phone: strings.TrimSpace(request.Phone)})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) ListCourses(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("level", c.Query("level"), []string{"licenciatura", "mestrado", "doutoramento", "tecnico_medio", "cet"}, "level inválido.")
	validationErrors.Enum("regime", c.Query("regime"), []string{"presencial", "distancia", "misto"}, "regime inválido.")
	validationErrors.UUID("university_id", c.Query("university_id"), "university_id deve ser um UUID válido.")
	validationErrors.Enum("sort", c.Query("sort"), []string{"name_asc", "name_desc", "created_at_asc"}, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListCourses(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, PaginationParams{Page: queryInt(c, "page", defaultPage), PerPage: queryInt(c, "per_page", defaultPerPage)}, CourseListFilters{Query: strings.TrimSpace(c.Query("q")), Area: strings.TrimSpace(c.Query("area")), Level: strings.TrimSpace(c.Query("level")), Regime: strings.TrimSpace(c.Query("regime")), UniversityID: strings.TrimSpace(c.Query("university_id")), Sort: strings.TrimSpace(c.Query("sort"))})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetCourse(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.GetCourse(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) CreateCourse(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request createCourseRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("university_id", request.UniversityID, "university_id é obrigatório.")
	validationErrors.UUID("university_id", request.UniversityID, "university_id deve ser um UUID válido.")
	validationErrors.Required("name", request.Name, "Nome é obrigatório.")
	validationErrors.Required("area", request.Area, "Área é obrigatória.")
	validationErrors.Required("level", request.Level, "Nível é obrigatório.")
	validationErrors.Required("regime", request.Regime, "Regime é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.CreateCourse(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateCourseInput{UniversityID: strings.TrimSpace(request.UniversityID), Name: strings.TrimSpace(request.Name), Area: strings.TrimSpace(request.Area), Level: strings.TrimSpace(request.Level), Regime: strings.TrimSpace(request.Regime), DurationYears: request.DurationYears, HasDurationYears: request.DurationYears > 0, AnnualFee: strings.TrimSpace(request.AnnualFee), EntryRequirements: strings.TrimSpace(request.EntryRequirements)})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handler) UpdateCourse(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request updateCourseRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	validationErrors.UUID("university_id", request.UniversityID, "university_id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.UpdateCourse(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateCourseInput{ID: strings.TrimSpace(c.Params("id")), UniversityID: strings.TrimSpace(request.UniversityID), Name: strings.TrimSpace(request.Name), Area: strings.TrimSpace(request.Area), Level: strings.TrimSpace(request.Level), Regime: strings.TrimSpace(request.Regime), DurationYears: request.DurationYears, HasDurationYears: request.DurationYears > 0, AnnualFee: strings.TrimSpace(request.AnnualFee), EntryRequirements: strings.TrimSpace(request.EntryRequirements)})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) ListOpportunities(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("type", c.Query("type"), []string{"bolsa", "estagio", "emprego", "intercambio", "workshop", "competicao"}, "type inválido.")
	validationErrors.Bool("verified", c.Query("verified"), "verified deve ser true ou false.")
	validationErrors.Bool("active", c.Query("active"), "active deve ser true ou false.")
	validationErrors.Enum("sort", c.Query("sort"), []string{"title_asc", "title_desc", "deadline_asc"}, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListOpportunities(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, PaginationParams{Page: queryInt(c, "page", defaultPage), PerPage: queryInt(c, "per_page", defaultPerPage)}, OpportunityListFilters{Query: strings.TrimSpace(c.Query("q")), Type: strings.TrimSpace(c.Query("type")), Verified: queryBool(c, "verified"), Active: queryBool(c, "active"), Sort: strings.TrimSpace(c.Query("sort"))})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetOpportunity(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	result, err := h.service.GetOpportunity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) UpdateArticle(c *fiber.Ctx) error {
	var request updateArticleRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.UpdateArticle(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateArticleInput{
		ID:             strings.TrimSpace(c.Params("id")),
		Title:          strings.TrimSpace(request.Title),
		Excerpt:        strings.TrimSpace(request.Excerpt),
		Content:        strings.TrimSpace(request.Content),
		ContentJSON:    request.ContentJSON,
		CoverImageURL:  strings.TrimSpace(request.CoverImageURL),
		Type:           strings.TrimSpace(request.Type),
		SourceName:     strings.TrimSpace(request.SourceName),
		SourceURL:      strings.TrimSpace(request.SourceURL),
		SEOTitle:       strings.TrimSpace(request.SEOTitle),
		SEODescription: strings.TrimSpace(request.SEODescription),
		IsFeatured:     request.IsFeatured,
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) UpdateOpportunity(c *fiber.Ctx) error {
	var request updateOpportunityRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	validationErrors.RFC3339("deadline", request.Deadline, "deadline deve estar em formato RFC3339.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para publicação CMS.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.UpdateOpportunity(c.UserContext(), Actor{UserID: currentUser.ID, Role: currentUser.Role}, CreateOpportunityInput{
		ID:           strings.TrimSpace(c.Params("id")),
		Title:        strings.TrimSpace(request.Title),
		Type:         strings.TrimSpace(request.Type),
		EntityName:   strings.TrimSpace(request.EntityName),
		Description:  strings.TrimSpace(request.Description),
		Requirements: strings.TrimSpace(request.Requirements),
		Deadline:     strings.TrimSpace(request.Deadline),
		ApplyURL:     strings.TrimSpace(request.ApplyURL),
		Country:      strings.TrimSpace(request.Country),
		Language:     strings.TrimSpace(request.Language),
		Area:         strings.TrimSpace(request.Area),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func handleError(err error) error {
	message := err.Error()
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Recurso CMS não encontrado.")
	}
	if errors.Is(err, ErrForbidden) {
		return apierror.Forbidden("Não tem permissões para gerir este recurso.")
	}
	if strings.Contains(message, "duplicate key value") {
		return apierror.Conflict("Já existe um recurso com este slug.")
	}
	if strings.Contains(message, "required") || strings.Contains(message, "invalid ") {
		return apierror.Validation("Dados inválidos para publicação CMS.", nil)
	}
	return err
}

func queryInt(c *fiber.Ctx, key string, fallback int) int {
	rawValue := strings.TrimSpace(c.Query(key))
	if rawValue == "" {
		return fallback
	}
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return fallback
	}
	return value
}

func queryBool(c *fiber.Ctx, key string) *bool {
	rawValue := strings.TrimSpace(strings.ToLower(c.Query(key)))
	if rawValue == "" {
		return nil
	}
	if rawValue == "true" {
		value := true
		return &value
	}
	if rawValue == "false" {
		value := false
		return &value
	}
	return nil
}
