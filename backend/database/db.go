package database

import (
	"fmt"
	"log"
	"os"

	"outcraftly/accounts/config"
	"outcraftly/accounts/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultProducts returns the canonical product registry.
// Stripe price IDs are read from environment variables so they can differ
// between dev (price_test_…) and production (price_live_…) without code changes.
func defaultProducts() []models.Product {
	// Read optional Stripe price ID for email-warmup.
	var emailWarmupPriceID *string
	if v := os.Getenv("STRIPE_EMAIL_WARMUP_PRICE_ID"); v != "" {
		emailWarmupPriceID = &v
	}

	return []models.Product{
		{Name: "cold_email", Description: "AI-powered cold email outreach and automation"},
		{Name: "linkedin", Description: "LinkedIn outreach and connection automation"},
		{
			Name:          "email-warmup",
			Description:   "Email inbox warm-up to improve deliverability and sender reputation",
			StripePriceID: emailWarmupPriceID,
			RedirectURLs: []string{
				"http://localhost:3000/callback",
				"https://warmup.outcraftly.com/callback",
			},
		},
	}
}

// DB is the shared GORM database instance.
var DB *gorm.DB

// Connect opens a connection to PostgreSQL and runs auto-migrations.
func Connect(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	if err = DB.AutoMigrate(
		&models.User{},
		&models.Workspace{},
		&models.WorkspaceMember{},
		&models.WorkspaceInvite{},
		&models.Product{},
		&models.Subscription{},
		&models.BillingCustomer{},
	); err != nil {
		log.Fatalf("❌ Auto-migration failed: %v", err)
	}

	seedProducts()
	log.Println("✅ Database connected and migrations applied")
}

// seedProducts upserts the default product registry on every startup.
//
// • New products (by name) are inserted.
// • Existing products have their description, redirect_urls, and
//   stripe_price_id brought in sync with the code/env values so that
//   deploy-time changes take effect without manual DB patches.
func seedProducts() {
	for _, p := range defaultProducts() {
		var existing models.Product
		if DB.Where("name = ?", p.Name).First(&existing).Error != nil {
			// ── INSERT ──────────────────────────────────────────────
			if err := DB.Create(&p).Error; err != nil {
				log.Printf("[seed] WARNING: could not insert product %q: %v", p.Name, err)
			} else {
				log.Printf("[seed] inserted product: %s", p.Name)
			}
			continue
		}

		// ── UPDATE existing row to match code/env values ─────────
		// Use the model struct so GORM's json serializer handles RedirectURLs.
		existing.Description = p.Description
		existing.RedirectURLs = p.RedirectURLs
		existing.IsActive = true
		// Only overwrite stripe_price_id when the env var is set;
		// this avoids blanking a price that was configured via admin API.
		if p.StripePriceID != nil {
			existing.StripePriceID = p.StripePriceID
		}

		if err := DB.Save(&existing).Error; err != nil {
			log.Printf("[seed] WARNING: could not update product %q: %v", p.Name, err)
		} else {
			log.Printf("[seed] synced product: %s", p.Name)
		}
	}

	// Deactivate the legacy "warmup" product if it exists — replaced by "email-warmup".
	if res := DB.Model(&models.Product{}).Where("name = ? AND is_active = true", "warmup").
		Update("is_active", false); res.RowsAffected > 0 {
		log.Println("[seed] deactivated legacy product: warmup (replaced by email-warmup)")
	}
}
