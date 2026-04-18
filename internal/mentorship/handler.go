package mentorship

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	appauth "github.com/rnrnshn/oportunidades-api/internal/auth"
	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
	"github.com/rnrnshn/oportunidades-api/pkg/validation"
)

type Handler struct {
	service *Service
}

type sessionRequest struct {
	MentorID    string `json:"mentor_id"`
	Message     string `json:"message"`
	ScheduledAt string `json:"scheduled_at"`
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListMentors(c *fiber.Ctx) error {
	result, err := h.service.ListMentors(c.UserContext(), PaginationParams{
		Page:    queryInt(c, "page", defaultPage),
		PerPage: queryInt(c, "per_page", defaultPerPage),
	})
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) GetMentorByID(c *fiber.Ctx) error {
	result, err := h.service.GetMentorByID(c.UserContext(), strings.TrimSpace(c.Params("id")))
	if err != nil {
		return handleError(err)
	}
	return c.JSON(result)
}

func (h *Handler) CreateSessionRequest(c *fiber.Ctx) error {
	currentUser, ok := appauth.CurrentUser(c)
	if !ok {
		return apierror.Unauthorized("Token inválido.")
	}

	var request sessionRequest
	if err := c.BodyParser(&request); err != nil {
		return apierror.Validation("Payload inválido.", nil)
	}
	validationErrors := validation.New()
	validationErrors.Required("mentor_id", request.MentorID, "mentor_id é obrigatório.")
	validationErrors.Required("message", request.Message, "message é obrigatório.")
	validationErrors.UUID("mentor_id", request.MentorID, "mentor_id deve ser um UUID válido.")
	validationErrors.RFC3339("scheduled_at", request.ScheduledAt, "scheduled_at deve estar em formato RFC3339.")
	if validationErrors.HasAny() {
		return apierror.Validation("Dados inválidos para pedir sessão de mentoria.", validationErrors.Details())
	}

	result, err := h.service.CreateSessionRequest(c.UserContext(), SessionRequestInput{
		MentorID:    strings.TrimSpace(request.MentorID),
		RequesterID: currentUser.ID,
		Message:     strings.TrimSpace(request.Message),
		ScheduledAt: strings.TrimSpace(request.ScheduledAt),
	})
	if err != nil {
		return handleError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(result)
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
		return apierror.NotFound("Mentor não encontrado.")
	}
	if strings.Contains(err.Error(), "requester cannot book own mentor profile") {
		return apierror.Validation("Dados inválidos para pedir sessão de mentoria.", nil)
	}
	return err
}
