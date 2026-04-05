package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

// GET /api/v1/config  (public)
// Returns non-secret client configuration — safe to expose to the browser.
// The Stripe publishable key is intentionally public (used in <stripe-pricing-table>).
func GetConfig(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"stripe_publishable_key": os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		"app_env":                os.Getenv("APP_ENV"),
	})
}
