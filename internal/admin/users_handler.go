package admin

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type updateUserRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) ListUsers(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.IntRange("page", c.Query("page"), 1, 100000, "page deve ser um inteiro >= 1.")
	validationErrors.IntRange("per_page", c.Query("per_page"), 1, 100, "per_page deve estar entre 1 e 100.")
	validationErrors.Enum("role", c.Query("role"), UserRoles, "role inválido.")
	validationErrors.Enum("sort", c.Query("sort"), []string{"name_asc", "name_desc", "created_at_asc"}, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListUsers(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, UserListFilters{
		Query: strings.TrimSpace(c.Query("q")),
		Role:  strings.TrimSpace(c.Query("role")),
		Sort:  strings.TrimSpace(c.Query("sort")),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.GetUser(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) UpdateUserRole(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	var request updateUserRoleRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	role := strings.TrimSpace(request.Role)
	validationErrors.Required("role", role, "role é obrigatório.")
	validationErrors.Enum("role", role, UserRoles, "role inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.UpdateUserRole(c.UserContext(), UpdateUserRoleInput{
		UserID:  strings.TrimSpace(c.Params("id")),
		Role:    role,
		ActorID: currentUser.ID,
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}
