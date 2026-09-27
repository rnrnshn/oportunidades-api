package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/rnrnshn/oportunidades-api/pkg/apierror"
)

func RateLimit(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() + "|" + c.Path()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return apierror.New(fiber.StatusTooManyRequests, "RATE_LIMITED", "Demasiadas tentativas. Tenta novamente dentro de alguns minutos.", nil)
		},
	})
}
