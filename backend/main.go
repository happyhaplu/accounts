package main

import (
"log"
"mime"
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

	// Pre-read index.html once at startup (small, never changes at runtime).
	indexHTML, err := os.ReadFile("./dist/index.html")
	if err != nil {
		log.Fatalf("❌ Cannot read ./dist/index.html: %v", err)
	}

	app.Get("/*", func(c *fiber.Ctx) error {
		// Build safe path inside dist/ (filepath.Clean resolves "..")
		clean := filepath.Clean(c.Path())
		fp := filepath.Join("dist", clean)

		// Try to serve the real file with explicit Content-Type.
		// We use os.ReadFile + c.Send instead of c.SendFile to guarantee
		// the correct MIME type is set — no reliance on fasthttp internals.
		if data, readErr := os.ReadFile(fp); readErr == nil {
			if ct := mime.TypeByExtension(filepath.Ext(fp)); ct != "" {
				c.Set("Content-Type", ct)
			}
			return c.Send(data)
		}

		// Missing file WITH a file extension (.js, .css, .png, …) → 404.
		// This prevents serving index.html (text/html) for a missing .js
		// asset, which browsers reject as a MIME-type mismatch.
		if filepath.Ext(clean) != "" {
			return c.SendStatus(fiber.StatusNotFound)
		}

		// No extension → Vue Router client-side route → serve index.html.
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(indexHTML)
	})
}

log.Printf("🚀 Gour Accounts API running on :%s", cfg.Port)
log.Fatal(app.Listen(":" + cfg.Port))
}
