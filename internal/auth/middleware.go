package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

const userContextKey = "auth_user"

type AuthenticatedUser struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

func RequireAuth(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawAuthorization := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
		if rawAuthorization == "" {
			return apierror.Unauthorized("Token em falta.")
		}

		parts := strings.SplitN(rawAuthorization, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			return apierror.Unauthorized("Token inválido.")
		}

		claims, err := service.ParseAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			return apierror.Unauthorized("Token inválido ou expirado.")
		}

		c.Locals(userContextKey, AuthenticatedUser{ID: claims.Subject, Role: claims.Role})
		return c.Next()
	}
}

func RequireRole(service *Service, roles ...string) fiber.Handler {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		if err := RequireAuth(service)(c); err != nil {
			return err
		}

		user, ok := c.Locals(userContextKey).(AuthenticatedUser)
		if !ok {
			return apierror.Unauthorized("Token inválido.")
		}

		if _, allowed := allowedRoles[user.Role]; !allowed {
			return apierror.Forbidden("Não tem permissões para esta acção.")
		}

		return c.Next()
	}
}

func CurrentUser(c *fiber.Ctx) (AuthenticatedUser, bool) {
	user, ok := c.Locals(userContextKey).(AuthenticatedUser)
	return user, ok
}
