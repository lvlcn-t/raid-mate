package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

// Context sets the user context on the request.
// The user context then contains the logger and open telemetry span.
func Context(ctx context.Context) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.SetUserContext(ctx)
		return c.Next()
	}
}
