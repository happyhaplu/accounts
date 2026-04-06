package main

import (
"log"
"os"
"path/filepath"
"strings"

"outcraftly/accounts/config"
"outcraftly/accounts/database"
"outcraftly/accounts/routes"

"github.com/gofiber/fiber/v2"
"github.com/gofiber/fiber/v2/middleware/cors"
"github.com/gofiber/fiber/v2/middleware/logger"
stripe "github.com/stripe/stripe-go/v76"
)

// webMIME maps file extensions to MIME types for static asset serving.
// We use a hardcoded map instead of Go's mime.TypeByExtension because
// the system MIME database on Alpine Linux is missing/broken — .js files
// were being served as text/html, causing browsers to reject them.
var webMIME = map[string]string{
".js":    "application/javascript; charset=utf-8",
".mjs":   "application/javascript; charset=utf-8",
".css":   "text/css; charset=utf-8",
".html":  "text/html; charset=utf-8",
".htm":   "text/html; charset=utf-8",
".json":  "application/json; charset=utf-8",
".svg":   "image/svg+xml",
".png":   "image/png",
".jpg":   "image/jpeg",
".jpeg":  "image/jpeg",
".gif":   "image/gif",
".ico":   "image/x-icon",
".webp":  "image/webp",
".avif":  "image/avif",
".woff":  "font/woff",
".woff2": "font/woff2",
".ttf":   "font/ttf",
".eot":   "application/vnd.ms-fontobject",
".map":   "application/json",
".txt":   "text/plain; charset=utf-8",
".xml":   "text/xml; charset=utf-8",
".wasm":  "application/wasm",
".pdf":   "application/pdf",
}

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

// ── Uploads — always registered (dev + production) ───────────────────────
// Product logos are written to ./uploads/logos/ by UploadProductLogo.
// This route must be outside the dist/ block so it works in local dev too
// (no dist/ exists in dev; Vite dev server proxies /uploads/* to here).
app.Get("/uploads/*", func(c *fiber.Ctx) error {
	urlPath := filepath.Clean(c.Path())
	uploadFilePath := filepath.Join(".", urlPath)
	if data, readErr := os.ReadFile(uploadFilePath); readErr == nil {
		ext := strings.ToLower(filepath.Ext(uploadFilePath))
		if ct, ok := webMIME[ext]; ok {
			c.Set("Content-Type", ct)
		} else {
			c.Set("Content-Type", "application/octet-stream")
		}
		c.Set("Cache-Control", "public, max-age=3600")
		return c.Send(data)
	}
	return c.SendStatus(fiber.StatusNotFound)
})

// ── SPA static file serving (production) ─────────────────────────────────
// In Docker the built Vue dist/ directory lives at ./dist alongside the
// binary.  Fiber serves it directly — no nginx needed.
// In local development this directory doesn't exist; the Vite dev server
// on :5173 serves the frontend instead, so this block is skipped.
if _, err := os.Stat("./dist/index.html"); err == nil {

	// Pre-read index.html once at startup (small, never changes at runtime).
	indexHTML, err := os.ReadFile("./dist/index.html")
	if err != nil {
		log.Fatalf("❌ Cannot read ./dist/index.html: %v", err)
	}

	// Startup verification — log what we're actually serving so we can
	// confirm in Coolify logs that the correct build was deployed.
	snippet := string(indexHTML)
	if len(snippet) > 500 {
		snippet = snippet[:500]
	}
	log.Printf("📁 Serving Vue SPA from ./dist\n--- index.html (first 500 chars) ---\n%s\n---", snippet)

	app.Get("/*", func(c *fiber.Ctx) error {
		urlPath := filepath.Clean(c.Path())
		fp := filepath.Join("dist", urlPath)

		// Try to serve the real file with explicit Content-Type.
		// We use os.ReadFile + c.Send instead of c.SendFile to guarantee
		// the correct MIME type — no reliance on fasthttp internals.
		if data, readErr := os.ReadFile(fp); readErr == nil {
			ext := strings.ToLower(filepath.Ext(fp))
			if ct, ok := webMIME[ext]; ok {
				c.Set("Content-Type", ct)
			} else {
				c.Set("Content-Type", "application/octet-stream")
			}

			// Hashed assets (Vite adds content-hash to filenames) are
			// immutable — let browsers cache them aggressively.
			if strings.HasPrefix(urlPath, "/assets/") {
				c.Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			return c.Send(data)
		}

		// Missing file WITH a file extension (.js, .css, .png, …) → 404.
		// This prevents serving index.html (text/html) for a missing .js
		// asset, which browsers reject as a MIME-type mismatch.
		if filepath.Ext(urlPath) != "" {
			return c.SendStatus(fiber.StatusNotFound)
		}

		// No extension → Vue Router client-side route → serve index.html.
		// no-cache ensures browsers always revalidate so they pick up new
		// deploys immediately (the JS/CSS have content hashes in filenames,
		// so those are cached aggressively — see above).
		c.Set("Content-Type", "text/html; charset=utf-8")
		c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
		return c.Send(indexHTML)
	})
}

log.Printf("🚀 Gour Accounts API running on :%s", cfg.Port)
log.Fatal(app.Listen(":" + cfg.Port))
}
