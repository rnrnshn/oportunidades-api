package middleware

import (
	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
)

func Recover() fiber.Handler {
	return recovermw.New()
}
