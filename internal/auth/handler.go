package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct {
	service *Service
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}
type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}
type verifyEmailRequest struct {
	Token string `json:"token"`
}
type deactivateAccountRequest struct {
	CurrentPassword string `json:"current_password"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var request registerRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}

	validationErrors := validation.New()
	validationErrors.Required("name", request.Name, "Nome é obrigatório.")
	validationErrors.Required("email", request.Email, "Email é obrigatório.")
	validationErrors.Required("password", request.Password, "Password é obrigatória.")
	validationErrors.MinLength("password", request.Password, 8, "Password deve ter pelo menos 8 caracteres.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}

	result, err := h.service.Register(c.UserContext(), RegisterInput{
		Email:    strings.TrimSpace(strings.ToLower(request.Email)),
		Password: request.Password,
		Name:     strings.TrimSpace(request.Name),
	})
	if err != nil {
		return handleServiceError(err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	return c.Status(fiber.StatusCreated).JSON(authSuccessResponse(result))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var request loginRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}

	validationErrors := validation.New()
	validationErrors.Required("email", request.Email, "Email é obrigatório.")
	validationErrors.Required("password", request.Password, "Password é obrigatória.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}

	result, err := h.service.Login(c.UserContext(), LoginInput{
		Email:    strings.TrimSpace(strings.ToLower(request.Email)),
		Password: request.Password,
	})
	if err != nil {
		return handleServiceError(err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	return c.JSON(authSuccessResponse(result))
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies(h.service.RefreshCookieName())
	if refreshToken == "" {
		return apierror.Unauthorized("Sessão inválida. Faça login novamente.")
	}

	result, err := h.service.Refresh(c.UserContext(), refreshToken)
	if err != nil {
		return handleServiceError(err)
	}

	h.setRefreshCookie(c, result.RefreshToken)
	return c.JSON(authSuccessResponse(result))
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	if err := h.service.Logout(c.UserContext(), c.Cookies(h.service.RefreshCookieName())); err != nil {
		return handleServiceError(err)
	}

	h.clearRefreshCookie(c)
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Sessão terminada com sucesso.",
		},
	})
}

func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	currentUser, ok := CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	if err := h.service.LogoutAll(c.UserContext(), currentUser.ID); err != nil {
		return handleServiceError(err)
	}
	h.clearRefreshCookie(c)
	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"message": "Todas as sessões foram terminadas com sucesso.",
		},
	})
}

func (h *Handler) ForgotPassword(c *fiber.Ctx) error {
	var request forgotPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("email", request.Email, "Email é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.ForgotPassword(c.UserContext(), ForgotPasswordInput{Email: strings.TrimSpace(strings.ToLower(request.Email))})
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(actionSuccessResponse(result))
}

func (h *Handler) ResetPassword(c *fiber.Ctx) error {
	var request resetPasswordRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("token", request.Token, "Token é obrigatório.")
	validationErrors.Required("new_password", request.NewPassword, "Nova password é obrigatória.")
	validationErrors.MinLength("new_password", request.NewPassword, 8, "A nova password deve ter pelo menos 8 caracteres.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.ResetPassword(c.UserContext(), ResetPasswordInput{Token: strings.TrimSpace(request.Token), NewPassword: strings.TrimSpace(request.NewPassword)})
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(actionSuccessResponse(result))
}

func (h *Handler) SendVerification(c *fiber.Ctx) error {
	currentUser, ok := CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	result, err := h.service.SendVerification(c.UserContext(), SendVerificationInput{UserID: currentUser.ID})
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(actionSuccessResponse(result))
}

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	var request verifyEmailRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("token", request.Token, "Token é obrigatório.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.VerifyEmail(c.UserContext(), VerifyEmailInput{Token: strings.TrimSpace(request.Token)})
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(actionSuccessResponse(result))
}

func (h *Handler) DeactivateAccount(c *fiber.Ctx) error {
	currentUser, ok := CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}
	var request deactivateAccountRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("current_password", request.CurrentPassword, "Password actual é obrigatória.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos.", validationErrors.Details())
	}
	result, err := h.service.DeactivateAccount(c.UserContext(), currentUser.ID, strings.TrimSpace(request.CurrentPassword))
	if err != nil {
		return handleServiceError(err)
	}
	h.clearRefreshCookie(c)
	return c.JSON(actionSuccessResponse(result))
}

func (h *Handler) setRefreshCookie(c *fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     h.service.RefreshCookieName(),
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   h.service.RefreshCookieSecure(),
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/",
		MaxAge:   h.service.RefreshCookieMaxAge(),
	})
}

func (h *Handler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     h.service.RefreshCookieName(),
		Value:    "",
		HTTPOnly: true,
		Secure:   h.service.RefreshCookieSecure(),
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/",
		Expires:  timeZero(),
		MaxAge:   -1,
	})
}

func authSuccessResponse(result *AuthResult) fiber.Map {
	return fiber.Map{
		"data": fiber.Map{
			"access_token": result.AccessToken,
			"expires_in":   result.ExpiresIn,
			"user":         result.User,
		},
	}
}

func actionSuccessResponse(result *ActionResult) fiber.Map {
	data := fiber.Map{"message": result.Message}
	if result.DebugToken != "" {
		data["debug_token"] = result.DebugToken
	}
	return fiber.Map{"data": data}
}

func handleServiceError(err error) error {
	if errors.Is(err, ErrInvalidCredentials) {
		return apierror.Unauthorized("Credenciais inválidas.")
	}

	if strings.Contains(err.Error(), "email already exists") {
		return apierror.Conflict("Já existe uma conta com este email.")
	}
	if strings.Contains(err.Error(), "new password must be at least 8 characters") {
		return apierror.Validation("Dados inválidos.", map[string]any{"fields": []map[string]string{{"field": "new_password", "reason": "min_length", "message": "A nova password deve ter pelo menos 8 caracteres."}}})
	}
	if strings.Contains(err.Error(), "current password is invalid") {
		return apierror.Unauthorized("Password actual inválida.")
	}

	return fmt.Errorf("auth handler: %w", err)
}
