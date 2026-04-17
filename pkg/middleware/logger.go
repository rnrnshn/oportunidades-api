package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func Logger(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()

		event := log.Info()
		if err != nil {
			event = log.Error().Err(err)
		}

		event.
			Str("method", c.Method()).
			Str("path", c.OriginalURL()).
			Int("status", c.Response().StatusCode()).
			Dur("duration", time.Since(startedAt)).
			Str("ip", c.IP()).
			Msg("http_request")

		return err
	}
}
