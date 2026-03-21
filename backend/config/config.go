package config

import (
"log"
"os"
"strings"

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

// ── Cross-app auth (Reach, Warmup, etc.) ──────────────────────
Environment      string   // "development" | "production"
AccountsBaseURL  string   // e.g. https://accounts.gour.io
JWTIssuer        string   // iss claim — accounts.gour.io
JWTAudience      string   // aud claim — reach
JWTPrivateKey    string   // RS256 PEM (optional, falls back to HS256)
JWTPublicKey     string   // RS256 PEM (optional)
JWKSURL          string   // /.well-known/jwks.json (optional)
AuthRedirectURIs []string // allowed redirect callback URIs

// Stripe — billing integration
StripeSecretKey      string
StripePublishableKey string
StripeWebhookSecret  string
}

// Cfg is the globally accessible configuration set by Load().
var Cfg *Config

// Load reads .env (if present) then maps env vars into a Config struct.
func Load() *Config {
if err := godotenv.Load(); err != nil {
log.Println("⚠  No .env file found — falling back to environment variables")
}

// CORS_ALLOWED_ORIGINS takes precedence over legacy ALLOW_ORIGINS.
allowOrigins := getEnv("CORS_ALLOWED_ORIGINS", "")
if allowOrigins == "" {
allowOrigins = getEnv("ALLOW_ORIGINS", "http://localhost:5173")
}

cfg := &Config{
Port:         getEnv("PORT", "8080"),
DBHost:       getEnv("DB_HOST", "localhost"),
DBPort:       getEnv("DB_PORT", "5432"),
DBUser:       getEnv("DB_USER", "postgres"),
DBPassword:   getEnv("DB_PASSWORD", ""),
DBName:       getEnv("DB_NAME", "gour_accounts"),
JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
AllowOrigins: allowOrigins,

Environment:      getEnv("APP_ENV", getEnv("ENV", "development")),
AccountsBaseURL:  getEnv("ACCOUNTS_BASE_URL", getEnv("APP_URL", "http://localhost:5173")),
JWTIssuer:        getEnv("JWT_ISSUER", "accounts.gour.io"),
JWTAudience:      getEnv("JWT_AUDIENCE", "reach"),
JWTPrivateKey:    getEnv("JWT_PRIVATE_KEY", ""),
JWTPublicKey:     getEnv("JWT_PUBLIC_KEY", ""),
JWKSURL:          getEnv("JWKS_URL", ""),
AuthRedirectURIs: splitCSV(getEnv("AUTH_REDIRECT_URIS", "")),

StripeSecretKey:      getEnv("STRIPE_SECRET_KEY", ""),
StripePublishableKey: getEnv("STRIPE_PUBLISHABLE_KEY", ""),
StripeWebhookSecret:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
}

Cfg = cfg
return cfg
}

func getEnv(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}

// splitCSV splits a comma-separated string into a trimmed slice.
func splitCSV(v string) []string {
if strings.TrimSpace(v) == "" {
return nil
}
parts := strings.Split(v, ",")
out := make([]string, 0, len(parts))
for _, p := range parts {
if p = strings.TrimSpace(p); p != "" {
out = append(out, p)
}
}
return out
}
