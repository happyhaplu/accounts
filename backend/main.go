package main

import (
"log"
"os"

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

	// Serve static assets (JS, CSS, images, fonts).
	app.Static("/", "./dist", fiber.Static{
		Compress: true,
	})

	// SPA fallback: any path that didn't match an API route or a static
	// file gets index.html so Vue Router can handle client-side routing.
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./dist/index.html")
	})
}

log.Printf("🚀 Gour Accounts API running on :%s", cfg.Port)
log.Fatal(app.Listen(":" + cfg.Port))
}
