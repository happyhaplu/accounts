package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a single product offering in the Gour registry.
// New products are added by inserting a row — no code changes required elsewhere.
type Product struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"                          json:"id"`
	Name         string    `gorm:"type:varchar(80);uniqueIndex;not null"          json:"name"`
	Description  string    `gorm:"type:text"                                     json:"description"`
	// RedirectURLs holds per-product callback URLs (JSON-encoded in a text column).
	// Index 0 is used as the product's home URL shown in the Billing page.
	RedirectURLs []string `gorm:"serializer:json;type:text;default:'[]'"          json:"redirect_urls"`
	IsActive     bool      `gorm:"not null;default:true"                          json:"is_active"`
	// LogoURL holds the relative path to the uploaded product logo, e.g. /uploads/logos/<id>.png
	// Empty string means no logo has been uploaded (UI falls back to letter initial).
	LogoURL      string    `gorm:"type:varchar(500);default:''"                   json:"logo_url"`
	// APIKey is the secret key given to each product for server-to-server calls.
	// Format: gour_ce_<32 hex chars>  (e.g. gour_ce_a1b2c3...)
	// Treat like a password — only shown in the Admin UI, never in public endpoints.
	APIKey       string    `gorm:"type:varchar(120);uniqueIndex"                  json:"api_key"`
	CreatedAt    time.Time `                                                      json:"created_at"`
}

// GenerateAPIKey creates a new random API key in the format gour_ce_<32 hex chars>.
func GenerateAPIKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generateAPIKey: %w", err)
	}
	return "gour_ce_" + hex.EncodeToString(b), nil
}

func (p *Product) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.APIKey == "" {
		key, err := GenerateAPIKey()
		if err != nil {
			return err
		}
		p.APIKey = key
	}
	return nil
}
