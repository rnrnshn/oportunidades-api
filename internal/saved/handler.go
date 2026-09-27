package saved

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	user, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros inválidos.", validationErrors.Details())
	}
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", strconv.Itoa(defaultPerPage)))
	result, err := h.service.List(c.UserContext(), user.ID, page, perPage)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) IDs(c *fiber.Ctx) error {
	user, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.IDs(c.UserContext(), user.ID)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) Save(c *fiber.Ctx) error {
	user, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.Save(c.UserContext(), user.ID, strings.TrimSpace(c.Params("opportunityId"))); err != nil {
		return handleError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Unsave(c *fiber.Ctx) error {
	user, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.Unsave(c.UserContext(), user.ID, strings.TrimSpace(c.Params("opportunityId"))); err != nil {
		return handleError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Oportunidade não encontrada.")
	}
	return err
}
