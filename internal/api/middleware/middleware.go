package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
)

// Context sets the user context on the request.
// The user context then contains the logger and open telemetry span.
func Context(ctx context.Context) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.SetUserContext(ctx)
		return c.Next()
	}
}

// Logger logs the request. It does not log health checks.
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		log := logger.FromContext(c.UserContext())
		if c.Path() != "/healthz" {
			log.InfoContext(c.Context(), "Request received", "ip", c.IP(), "method", c.Method(), "path", c.Path())
		}
		return c.Next()
	}
}

// Recover recovers from panics and logs the error.
func Recover() fiber.Handler {
	return func(c fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log := logger.FromContext(c.UserContext())
				log.ErrorContext(c.Context(), "Panic recovered", "error", r)
				err = errors.Join(err, c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fiber.Map{
					"message": fmt.Sprintf("panic: %v", r),
					"code":    http.StatusInternalServerError,
				}}))
			}
		}()
		return c.Next()
	}
}

// Authenticate checks if the request is authenticated.
func Authenticate() fiber.Handler {
	// TODO: Implement authentication logic.
	return func(c fiber.Ctx) error {
		if c.Get("Authorization") == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": fiber.Map{
				"message": "unauthorized",
				"code":    http.StatusUnauthorized,
			}})
		}
		return c.Next()
	}
}

// Authorize checks if the request is authorized.
func Authorize() fiber.Handler {
	// TODO: Implement authorization logic.
	return func(c fiber.Ctx) error {
		if c.Get("Authorization") != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": fiber.Map{
				"message": "forbidden",
				"code":    http.StatusForbidden,
			}})
		}
		return c.Next()
	}
}
