package reports

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct{ service *Service }

type createReportRequest struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Reason     string `json:"reason"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request createReportRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("entity_type", request.EntityType, "entity_type é obrigatório.")
	validationErrors.Required("entity_id", request.EntityID, "entity_id é obrigatório.")
	validationErrors.Required("reason", request.Reason, "reason é obrigatório.")
	validationErrors.UUID("entity_id", request.EntityID, "entity_id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para denúncia.", validationErrors.Details())
	}
	result, err := h.service.Create(c.UserContext(), CreateReportInput{
		ReporterID: currentUser.ID,
		EntityType: strings.TrimSpace(request.EntityType),
		EntityID:   strings.TrimSpace(request.EntityID),
		Reason:     strings.TrimSpace(request.Reason),
	})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func handleError(err error) error {
	if strings.Contains(err.Error(), "invalid ") {
		return apierror.Validation("Dados inválidos para denúncia.", nil)
	}
	return err
}
