package apierror

import "github.com/gofiber/fiber/v2"

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Response struct {
	Error Error `json:"error"`
}

func New(status int, code string, message string, details map[string]any) error {
	return fiber.NewError(status, messageWithCode(code, message, details))
}

func Validation(message string, details map[string]any) error {
	return New(fiber.StatusBadRequest, "VALIDATION_ERROR", message, details)
}

func Unauthorized(message string) error {
	return New(fiber.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(message string) error {
	return New(fiber.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(message string) error {
	return New(fiber.StatusNotFound, "NOT_FOUND", message, nil)
}

func Conflict(message string) error {
	return New(fiber.StatusConflict, "CONFLICT", message, nil)
}

func Internal() error {
	return New(fiber.StatusInternalServerError, "INTERNAL_ERROR", "Ocorreu um erro. Tente novamente.", nil)
}

func Handler(c *fiber.Ctx, err error) error {
	fiberErr, ok := err.(*fiber.Error)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(Response{
			Error: Error{
				Code:    "INTERNAL_ERROR",
				Message: "Ocorreu um erro. Tente novamente.",
			},
		})
	}

	code, message, details := parseMessage(fiberErr.Message)
	if code == "" {
		code = "INTERNAL_ERROR"
	}

	return c.Status(fiberErr.Code).JSON(Response{
		Error: Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
