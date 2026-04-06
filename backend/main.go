package main

import (
"log"
"os"
"path/filepath"

"outcraftly/accounts/config"
"outcraftly/accounts/database"
"outcraftly/accounts/routes"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/cors"
"github.com/gofiber/fiber/v2/middleware/logger"
stripe "github.com/stripe/stripe-go/v76"
)

func main() {
// Load config (reads .env)
cfg := config.Load()

// Initialise Stripe (billing) — safe to set empty key, handlers check it.
stripe.Key = cfg.StripeSecretKey

// Connect to PostgreSQL and run auto-migrations
database.Connect(cfg)

// Fiber app
app := fiber.New(fiber.Config{
AppName: "Gour Accounts API v1.0",
})

// Global middleware
app.Use(logger.New())

// CORS — origins are resolved dynamically from the product registry in the
// database (each product's redirect_urls). Static origins (the Accounts
// frontend, dev localhost) come from the ALLOW_ORIGINS env var.
database.SetStaticOrigins(cfg.AllowOrigins)
app.Use(cors.New(cors.Config{
	AllowOriginsFunc: database.IsAllowedOrigin,
	AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-API-Key, X-Admin-Secret",
	AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
}))

// Register all routes
routes.Setup(app)

// ── SPA static file serving (production) ─────────────────────────────────
// In Docker the built Vue dist/ directory lives at ./dist alongside the
// binary.  Fiber serves it directly — no nginx needed.
// In local development this directory doesn't exist; the Vite dev server
// on :5173 serves the frontend instead, so this block is skipped.
if _, err := os.Stat("./dist/index.html"); err == nil {
	log.Println("📁 Serving Vue SPA from ./dist")

	// Single catch-all: serve the real file when it exists in ./dist,
	// otherwise return index.html for Vue Router client-side routing.
	// c.SendFile sets Content-Type from the extension (.js → application/
	// javascript, .css → text/css, …) so MIME-type mismatches are impossible.
	app.Get("/*", func(c *fiber.Ctx) error {
		fp := filepath.Join("./dist", filepath.Clean(c.Path()))
		if info, err := os.Stat(fp); err == nil && !info.IsDir() {
			return c.SendFile(fp)
		}
		return c.SendFile("./dist/index.html")
	})
}

log.Printf("🚀 Gour Accounts API running on :%s", cfg.Port)
log.Fatal(app.Listen(":" + cfg.Port))
}
