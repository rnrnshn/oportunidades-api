package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func SecurityHeaders() fiber.Handler {
	return helmet.New(helmet.Config{
		ContentSecurityPolicy:     "default-src 'none'; frame-ancestors 'none'",
		CrossOriginResourcePolicy: "same-site",
	})
}
