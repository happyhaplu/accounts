package database

import (
	"fmt"
	"log"

	"outcraftly/accounts/config"
	"outcraftly/accounts/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// defaultProducts are seeded on every startup (idempotent — only inserted if missing).
var defaultProducts = []models.Product{
	{Name: "cold_email",  Description: "AI-powered cold email outreach and automation"},
	{Name: "linkedin",    Description: "LinkedIn outreach and connection automation"},
	{Name: "warmup",      Description: "Email inbox warm-up to improve deliverability"},
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

// seedProducts inserts default products if they do not already exist.
// It is safe to call on every startup — existing rows are never modified.
func seedProducts() {
	for _, p := range defaultProducts {
		var existing models.Product
		if DB.Where("name = ?", p.Name).First(&existing).Error != nil {
			if err := DB.Create(&p).Error; err != nil {
				log.Printf("[seed] WARNING: could not insert product %q: %v", p.Name, err)
			} else {
				log.Printf("[seed] inserted product: %s", p.Name)
			}
		}
	}
}
