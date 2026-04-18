package account

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct{ service *Service }

type updateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) GetMe(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.GetProfile(c.UserContext(), currentUser.ID)
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) UpdateMe(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request updateProfileRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("name", request.Name, "Nome é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.UpdateProfile(c.UserContext(), UpdateProfileInput{
		UserID:    currentUser.ID,
		Name:      strings.TrimSpace(request.Name),
		AvatarURL: strings.TrimSpace(request.AvatarURL),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request changePasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("current_password", request.CurrentPassword, "Password actual é obrigatória.")
	validationErrors.Required("new_password", request.NewPassword, "Nova password é obrigatória.")
	validationErrors.MinLength("new_password", request.NewPassword, 8, "A nova password deve ter pelo menos 8 caracteres.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.ChangePassword(c.UserContext(), ChangePasswordInput{
		UserID:          currentUser.ID,
		CurrentPassword: strings.TrimSpace(request.CurrentPassword),
		NewPassword:     strings.TrimSpace(request.NewPassword),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Conta não encontrada.")
	}
	if strings.Contains(err.Error(), "current and new passwords are required") {
		return apierror.Validation("Dados inválidos.", nil)
	}
	if strings.Contains(err.Error(), "current password is invalid") {
		return apierror.Unauthorized("Password actual inválida.")
	}
	return err
}
