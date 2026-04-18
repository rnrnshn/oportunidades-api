package catalog

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

func (h *Handler) ListUniversities(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("type", c.Query("type"), []string{"publica", "privada", "instituto", "academia"}, "type deve ser publica, privada, instituto ou academia.")
	validationErrors.Bool("verified", c.Query("verified"), "verified deve ser true ou false.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListUniversities(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, UniversityFilters{
		Query:    strings.TrimSpace(c.Query("q")),
		Province: strings.TrimSpace(c.Query("province")),
		Type:     strings.TrimSpace(c.Query("type")),
		Verified: queryBool(c, "verified"),
	})
	if err != nil {
		return handleError(err)
	}

	return c.JSON(result)
}

func (h *Handler) GetUniversityBySlug(c *fiber.Ctx) error {
	result, err := h.service.GetUniversityBySlug(c.UserContext(), strings.TrimSpace(c.Params("slug")))
	if err != nil {
		return handleError(err)
	}

	return c.JSON(result)
}

func (h *Handler) ListCourses(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("level", c.Query("level"), []string{"licenciatura", "mestrado", "doutoramento", "tecnico_medio", "cet"}, "level inválido.")
	validationErrors.Enum("regime", c.Query("regime"), []string{"presencial", "distancia", "misto"}, "regime inválido.")
	validationErrors.UUID("university_id", c.Query("university_id"), "university_id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListCourses(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, CourseFilters{
		Query:        strings.TrimSpace(c.Query("q")),
		Area:         strings.TrimSpace(c.Query("area")),
		Level:        strings.TrimSpace(c.Query("level")),
		Regime:       strings.TrimSpace(c.Query("regime")),
		Province:     strings.TrimSpace(c.Query("province")),
		UniversityID: strings.TrimSpace(c.Query("university_id")),
	})
	if err != nil {
		return handleError(err)
	}

	return c.JSON(result)
}

func (h *Handler) GetCourseBySlug(c *fiber.Ctx) error {
	result, err := h.service.GetCourseBySlug(c.UserContext(), strings.TrimSpace(c.Params("slug")))
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
