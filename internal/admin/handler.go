package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct{ service *Service }

type updateReportStatusRequest struct {
	Status string `json:"status"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublishArticle(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.PublishArticle(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) VerifyOpportunity(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.VerifyOpportunity(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) ListReports(c *fiber.Ctx) error {
	result, err := h.service.ListReports(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) UpdateReportStatus(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.Required("id", c.Params("id"), "id é obrigatório.")
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	var request updateReportStatusRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors.Required("status", request.Status, "status é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.UpdateReportStatus(c.UserContext(), UpdateReportStatusInput{ReportID: strings.TrimSpace(c.Params("id")), Status: strings.TrimSpace(request.Status)})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
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

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Recurso não encontrado.")
	}
	return err
}
