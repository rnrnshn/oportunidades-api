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

func handleServiceError(err error) error {
	if errors.Is(err, ErrInvalidCredentials) {
		return apierror.Unauthorized("Credenciais inválidas.")
	}

	if strings.Contains(err.Error(), "email already exists") {
		return apierror.Conflict("Já existe uma conta com este email.")
	}

	return fmt.Errorf("auth handler: %w", err)
}
