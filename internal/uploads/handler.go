package uploads

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct {
	service *Service
}

type presignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Folder      string `json:"folder"`
}

type confirmRequest struct {
	Path string `json:"path"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Presign(c *fiber.Ctx) error {
	var request presignRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("filename", request.Filename, "filename é obrigatório.")
	validationErrors.Required("content_type", request.ContentType, "content_type é obrigatório.")
	validationErrors.Required("folder", request.Folder, "folder é obrigatório.")
	validationErrors.Enum("folder", request.Folder, []string{"articles", "opportunities", "universities", "users"}, "folder inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para upload.", validationErrors.Details())
	}
	result, err := h.service.Presign(c.UserContext(), PresignInput{Filename: strings.TrimSpace(request.Filename), ContentType: strings.TrimSpace(request.ContentType), Folder: strings.TrimSpace(request.Folder)})
	if err != nil {
		return err
	}
	return c.JSON(result)
}

func (h *Handler) Confirm(c *fiber.Ctx) error {
	var request confirmRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("path", request.Path, "path é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para upload.", validationErrors.Details())
	}
	result, err := h.service.Confirm(c.UserContext(), ConfirmInput{Path: request.Path})
	if err != nil {
		return err
	}
	return c.JSON(result)
}
