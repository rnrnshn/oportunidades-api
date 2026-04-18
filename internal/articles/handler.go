package articles

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListArticles(c *fiber.Ctx) error {
	result, err := h.service.ListArticles(c.UserContext(), PaginationParams{
		Page: queryInt(c, "page", defaultPage), PerPage: queryInt(c, "per_page", defaultPerPage),
	}, Filters{
		Query:    strings.TrimSpace(c.Query("q")),
		Type:     strings.TrimSpace(c.Query("type")),
		Featured: queryBool(c, "featured"),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetArticleBySlug(c *fiber.Ctx) error {
	result, err := h.service.GetArticleBySlug(c.UserContext(), strings.TrimSpace(c.Params("slug")))
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
		v := true
		return &v
	}
	if rawValue == "false" {
		v := false
		return &v
	}
	return nil
}

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Artigo não encontrado.")
	}
	return err
}
