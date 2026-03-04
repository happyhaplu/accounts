package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// AdminOnly is a Fiber middleware that checks for a valid X-Admin-Secret header.
// The expected secret is read from the ADMIN_SECRET environment variable.
// This provides a simple, low-ceremony way to protect admin-only endpoints
// without building a full role system at MVP stage.
func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		secret := os.Getenv("ADMIN_SECRET")
		if secret == "" {
			// Guard against accidentally leaving the admin API open when the
			// env var is not configured.
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Admin API is not configured on this server",
			})
		}
		if c.Get("X-Admin-Secret") != secret {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Invalid or missing admin secret",
			})
		}
		return c.Next()
	}
}
