package main

import (
"log"

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
AppName: "Outcraftly Accounts API v1.0",
})

// Global middleware
app.Use(logger.New())
app.Use(cors.New(cors.Config{
AllowOrigins: cfg.AllowOrigins,
AllowHeaders: "Origin, Content-Type, Accept, Authorization",
AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
}))

// Register all routes
routes.Setup(app)

log.Printf("🚀 Outcraftly Accounts API running on :%s", cfg.Port)
log.Fatal(app.Listen(":" + cfg.Port))
}
