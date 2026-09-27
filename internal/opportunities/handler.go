package opportunities

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListOpportunities(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("type", c.Query("type"), []string{"bolsa", "estagio", "emprego", "intercambio", "workshop", "competicao"}, "type inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListOpportunities(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, Filters{
		Query:   strings.TrimSpace(c.Query("q")),
		Type:    strings.TrimSpace(c.Query("type")),
		Area:    strings.TrimSpace(c.Query("area")),
		Country: strings.TrimSpace(c.Query("country")),
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

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Recurso não encontrado.")
	}

	return err
}
