package config

import (
"log"
"os"

"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
Port         string
DBHost       string
DBPort       string
DBUser       string
DBPassword   string
DBName       string
JWTSecret    string
AllowOrigins string

// Stripe — billing integration
StripeSecretKey      string
StripePublishableKey string
StripeWebhookSecret  string
}

// Load reads .env (if present) then maps env vars into a Config struct.
func Load() *Config {
if err := godotenv.Load(); err != nil {
log.Println("⚠  No .env file found — falling back to environment variables")
}

return &Config{
Port:         getEnv("PORT", "8080"),
DBHost:       getEnv("DB_HOST", "localhost"),
DBPort:       getEnv("DB_PORT", "5432"),
DBUser:       getEnv("DB_USER", "postgres"),
DBPassword:   getEnv("DB_PASSWORD", ""),
DBName:       getEnv("DB_NAME", "outcraftly_accounts"),
JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
AllowOrigins: getEnv("ALLOW_ORIGINS", "http://localhost:5173"),

StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
}
}

func getEnv(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}
