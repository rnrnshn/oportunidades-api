package admin

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) PublishArticle(c *fiber.Ctx) error {
	result, err := h.service.PublishArticle(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) VerifyOpportunity(c *fiber.Ctx) error {
	result, err := h.service.VerifyOpportunity(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Recurso não encontrado.")
	}
	return err
}
