package database

import (
	"fmt"
	"log"
	"time"

	"outcraftly/accounts/config"
	"outcraftly/accounts/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the shared GORM database instance.
var DB *gorm.DB

// Connect opens a connection to PostgreSQL and runs auto-migrations.
func Connect(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	// Retry loop — postgres may pass its healthcheck (pg_isready) before the
	// user/database have finished initialising.  Retry for up to 30 seconds.
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("⏳ DB not ready (attempt %d/10): %v — retrying in 3 s...", attempt, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL after 10 attempts: %v", err)
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

	// Cleanup for removed billing-plan architecture.
	if err := DB.Exec(`DROP TABLE IF EXISTS product_plans`).Error; err != nil {
		log.Fatalf("❌ Failed to drop legacy table product_plans: %v", err)
	}

	ensureProductAPIKeys()
	log.Println("✅ Database connected and migrations applied")
}

// ensureProductAPIKeys back-fills api_key for rows created before this field existed.
func ensureProductAPIKeys() {
	var needsKey []models.Product
	DB.Where("api_key = '' OR api_key IS NULL").Find(&needsKey)
	for _, p := range needsKey {
		key, err := models.GenerateAPIKey()
		if err != nil {
			log.Printf("[seed] WARNING: could not generate api_key for %q: %v", p.Name, err)
			continue
		}
		if err := DB.Model(&p).Update("api_key", key).Error; err != nil {
			log.Printf("[seed] WARNING: could not save api_key for %q: %v", p.Name, err)
		} else {
			log.Printf("[seed] generated api_key for existing product: %s", p.Name)
		}
	}
}
