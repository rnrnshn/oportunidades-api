package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	corsmw "github.com/gofiber/fiber/v2/middleware/cors"
)

func CORS(allowedOrigins string) fiber.Handler {
	origins := strings.TrimSpace(allowedOrigins)
	wildcard := origins == "" || strings.Contains(origins, "*")
	if wildcard {
		origins = "*"
	}

	return corsmw.New(corsmw.Config{
		AllowOrigins: origins,
		// Browsers reject credentialed responses for "*", and allowing them would let any site use a visitor's session.
		AllowCredentials: !wildcard,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	})
}
