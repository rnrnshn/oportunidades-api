package opportunities

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListOpportunities(c *fiber.Ctx) error {
	result, err := h.service.ListOpportunities(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, Filters{
		Query:    strings.TrimSpace(c.Query("q")),
		Type:     strings.TrimSpace(c.Query("type")),
		Area:     strings.TrimSpace(c.Query("area")),
		Country:  strings.TrimSpace(c.Query("country")),
		Active:   queryBool(c, "active"),
		Verified: queryBool(c, "verified"),
	})
	if err != nil {
		return handleError(err)
	}

	return c.JSON(result)
}

func (h *Handler) GetOpportunityBySlug(c *fiber.Ctx) error {
	result, err := h.service.GetOpportunityBySlug(c.UserContext(), strings.TrimSpace(c.Params("slug")))
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

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Recurso não encontrado.")
	}

	return err
}
