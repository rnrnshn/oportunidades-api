package account

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

type Handler struct{ service *Service }

type updateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
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

func handleError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return apierror.NotFound("Conta não encontrada.")
	}
	if strings.Contains(err.Error(), "name is required") {
		return apierror.Validation("Nome é obrigatório.", nil)
	}
	return err
}
