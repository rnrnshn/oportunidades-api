package cms

import (
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
	CoverImageURL  string `json:"cover_image_url"`
	Type           string `json:"type"`
	SourceName     string `json:"source_name"`
	SourceURL      string `json:"source_url"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	IsFeatured     bool   `json:"is_featured"`
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

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

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
	result, err := h.service.CreateArticle(c.UserContext(), CreateArticleInput{
		AuthorID: currentUser.ID,
		Title:    strings.TrimSpace(request.Title), Excerpt: strings.TrimSpace(request.Excerpt), Content: strings.TrimSpace(request.Content), CoverImageURL: strings.TrimSpace(request.CoverImageURL),
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
	result, err := h.service.CreateOpportunity(c.UserContext(), CreateOpportunityInput{
		PublishedBy: currentUser.ID,
		Title:       strings.TrimSpace(request.Title), Type: strings.TrimSpace(request.Type), EntityName: strings.TrimSpace(request.EntityName), Description: strings.TrimSpace(request.Description), Requirements: strings.TrimSpace(request.Requirements), Deadline: strings.TrimSpace(request.Deadline), ApplyURL: strings.TrimSpace(request.ApplyURL), Country: strings.TrimSpace(request.Country), Language: strings.TrimSpace(request.Language), Area: strings.TrimSpace(request.Area),
	})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func handleError(err error) error {
	message := err.Error()
	if strings.Contains(message, "duplicate key value") {
		return apierror.Conflict("Já existe um recurso com este slug.")
	}
	if strings.Contains(message, "required") || strings.Contains(message, "invalid ") {
		return apierror.Validation("Dados inválidos para publicação CMS.", nil)
	}
	return err
}
