package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a single product offering in the Gour registry.
// New products are added by inserting a row — no code changes required elsewhere.
type Product struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"                          json:"id"`
	Name           string    `gorm:"type:varchar(80);uniqueIndex;not null"          json:"name"`
	Description    string    `gorm:"type:text"                                     json:"description"`
	// StripePriceID is the Stripe Price ID (price_xxx) for this product's default plan.
	// Set via admin PATCH /api/v1/admin/products/:id. Required for checkout button to appear.
	StripePriceID  *string  `gorm:"type:varchar(200)"                              json:"stripe_price_id,omitempty"`
	// RedirectURLs holds per-product callback URLs (JSON-encoded in a text column).
	// Index 0 is used as a fallback; production URLs start with "https://", dev with "http://localhost".
	RedirectURLs   []string `gorm:"serializer:json;type:text;default:'[]'"         json:"redirect_urls"`
	IsActive       bool     `gorm:"not null;default:true"                          json:"is_active"`
	CreatedAt      time.Time `                                                     json:"created_at"`
}

func (p *Product) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
