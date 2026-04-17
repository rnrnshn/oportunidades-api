package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	corsmw "github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(allowedOrigins string) fiber.Handler {
	origins := strings.TrimSpace(allowedOrigins)
	if origins == "" {
		origins = "*"
	}

	return corsmw.New(corsmw.Config{
		AllowOrigins: origins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	})
}
