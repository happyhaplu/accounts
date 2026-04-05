package handlers

import (
	"os"
	"strings"

	"outcraftly/accounts/config"

	"github.com/gofiber/fiber/v2"
)

// AdminLogin validates admin credentials and returns the ADMIN_SECRET used to
// protect all /api/v1/admin/* routes via the X-Admin-Secret header.
//
// POST /api/v1/admin/auth/login
// Body: { "email": "...", "password": "..." }
//
// Expected env vars:
//   ADMIN_EMAIL    — the admin's login email
//   ADMIN_PASSWORD — the admin's login password
//   ADMIN_SECRET   — the X-Admin-Secret value that gates admin routes
func AdminLogin(c *fiber.Ctx) error {
	type body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	req := new(body)
	if err := c.BodyParser(req); err != nil {
		return badRequest(c, "Invalid request body")
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	expectedEmail    := strings.ToLower(strings.TrimSpace(config.Cfg.AdminEmail))
	expectedPassword := config.Cfg.AdminPassword

	if expectedEmail == "" || expectedPassword == "" {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(errJSON("Admin login is not configured on this server"))
	}

	if req.Email != expectedEmail || req.Password != expectedPassword {
		return c.Status(fiber.StatusUnauthorized).
			JSON(errJSON("Invalid admin credentials"))
	}

	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		return c.Status(fiber.StatusServiceUnavailable).
			JSON(errJSON("Admin secret is not configured"))
	}

	return c.JSON(fiber.Map{
		"admin_secret": adminSecret,
		"message":      "Login successful",
	})
}
