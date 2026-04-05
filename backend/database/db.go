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

	// Cleanup for removed billing-plan architecture.
	if err := DB.Exec(`DROP TABLE IF EXISTS product_plans`).Error; err != nil {
		log.Fatalf("❌ Failed to drop legacy table product_plans: %v", err)
	}

	if cfg.SeedDefaultProducts {
		seedDefaultProducts()
	}
	ensureProductAPIKeys()
	log.Println("✅ Database connected and migrations applied")
}

// seedDefaultProducts inserts a small starter registry for local/dev usage.
// Production should keep this disabled and manage products from Admin UI.
func seedDefaultProducts() {
	defaults := []models.Product{
		{
			Name:        "email-warmup",
			Description: "Email inbox warm-up to improve deliverability and sender reputation",
			RedirectURLs: []string{
				"http://localhost:3000/callback",
				"https://warmup.gour.io/callback",
			},
		},
		{
			Name:        "reach",
			Description: "LinkedIn automation and outreach — find leads, send connection requests, and manage campaigns at scale",
			RedirectURLs: []string{
				"http://localhost:4000/auth/callback",
				"https://reach.gour.io/auth/callback",
			},
		},
		{
			Name:        "sendflow",
			Description: "Multi-channel cold outreach — email sequencing, sender rotation, and deliverability management",
			RedirectURLs: []string{
				"http://localhost:3000/callback",
				"https://sendflow.gour.io/callback",
			},
		},
	}

	for _, p := range defaults {
		var existing models.Product
		if DB.Where("name = ?", p.Name).First(&existing).Error == nil {
			continue
		}
		if err := DB.Create(&p).Error; err != nil {
			log.Printf("[seed] WARNING: could not insert product %q: %v", p.Name, err)
		} else {
			log.Printf("[seed] inserted product: %s", p.Name)
		}
	}
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
