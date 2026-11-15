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
	validationErrors.Enum("scope", c.Query("scope"), UserScopes, "scope inválido.")
	validationErrors.Enum("status", c.Query("status"), UserStatus, "status inválido.")
	validationErrors.Enum("sort", c.Query("sort"), UserSorts, "sort inválido.")
	if validationErrors.HasAny() {
		return apierror.Validation("Parâmetros de pesquisa inválidos.", validationErrors.Details())
	}
	result, err := h.service.ListUsers(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	}, UserListFilters{
		Query:  strings.TrimSpace(c.Query("q")),
		Scope:  strings.TrimSpace(c.Query("scope")),
		Role:   strings.TrimSpace(c.Query("role")),
		Status: strings.TrimSpace(c.Query("status")),
		Sort:   strings.TrimSpace(c.Query("sort")),
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

type suspendUserRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) SuspendUser(c *fiber.Ctx) error {
	validationErrors := validation.New()
	validationErrors.UUID("id", c.Params("id"), "id deve ser um UUID válido.")
	var request suspendUserRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors.Required("reason", strings.TrimSpace(request.Reason), "Indique o motivo da suspensão.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.SuspendUser(c.UserContext(), SuspendUserInput{UserID: strings.TrimSpace(c.Params("id")), Reason: request.Reason, ActorID: currentUser.ID}); err != nil {
		return handleError(err)
	}
	return h.GetUser(c)
}

func (h *Handler) UnsuspendUser(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.UnsuspendUser(c.UserContext(), strings.TrimSpace(c.Params("id")), currentUser.ID); err != nil {
		return handleError(err)
	}
	return h.GetUser(c)
}

type createInviteRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

func (h *Handler) ListInvites(c *fiber.Ctx) error {
	result, err := h.service.ListInvites(c.UserContext())
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) CreateInvite(c *fiber.Ctx) error {
	var request createInviteRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("email", strings.TrimSpace(request.Email), "Indique o email.")
	validationErrors.Required("name", strings.TrimSpace(request.Name), "Indique o nome.")
	validationErrors.Enum("role", request.Role, InviteRoles, "A função tem de ser administrador ou parceiro de conteúdo.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.CreateInvite(c.UserContext(), CreateInviteInput{Email: request.Email, Name: request.Name, Role: request.Role, ActorID: currentUser.ID})
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (h *Handler) ResendInvite(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.ResendInvite(c.UserContext(), strings.TrimSpace(c.Params("id")), currentUser.ID)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) RevokeInvite(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.RevokeInvite(c.UserContext(), strings.TrimSpace(c.Params("id")), currentUser.ID); err != nil {
		return handleError(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// Public: the token in the path is the credential.
func (h *Handler) PublicInvite(c *fiber.Ctx) error {
	result, err := h.service.PublicInvite(c.UserContext(), c.Params("token"))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

type acceptInviteRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (h *Handler) AcceptInvite(c *fiber.Ctx) error {
	var request acceptInviteRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	if len(request.Password) < 8 {
		return apierror.Validation("Dados inválidos.", map[string]any{"fields": []map[string]string{{"field": "password", "reason": "min_length", "message": "A password deve ter pelo menos 8 caracteres."}}})
	}
	result, err := h.service.AcceptInvite(c.UserContext(), c.Params("token"), request.Name, request.Password)
	if err != nil {
		return handleError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(result)
}
