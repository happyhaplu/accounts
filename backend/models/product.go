package models

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringArray is a []string that persists as a JSON array in a text column.
// It implements driver.Valuer and sql.Scanner so GORM uses it directly
// without the serializer:json tag.
//
// The Scan method gracefully handles corrupted or legacy non-JSON values
// (e.g. a raw URL string) by returning an empty slice instead of erroring,
// so a bad DB value never causes a 404 or breaks the list endpoint.
type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	b, err := json.Marshal([]string(s))
	if err != nil {
		return "[]", nil
	}
	return string(b), nil
}

func (s *StringArray) Scan(src interface{}) error {
	if src == nil {
		*s = StringArray{}
		return nil
	}
	var str string
	switch v := src.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		*s = StringArray{}
		return nil
	}
	str = strings.TrimSpace(str)
	if str == "" || str == "null" {
		*s = StringArray{}
		return nil
	}
	var result []string
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		// Graceful fallback: corrupted or legacy non-JSON value in DB.
		// Return empty slice instead of propagating an error that would
		// cause every query touching this product to fail.
		*s = StringArray{}
		return nil
	}
	*s = StringArray(result)
	return nil
}

// Product represents a single product offering in the Gour registry.
// New products are added by inserting a row — no code changes required elsewhere.
type Product struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"                          json:"id"`
	Name         string    `gorm:"type:varchar(80);uniqueIndex;not null"          json:"name"`
	Description  string    `gorm:"type:text"                                     json:"description"`
	// RedirectURLs holds per-product callback URLs stored as a JSON array in a text column.
	// Index 0 is used as the product's home URL shown in the Billing page.
	// StringArray implements driver.Valuer/sql.Scanner and handles corrupted values gracefully.
	RedirectURLs StringArray `gorm:"type:text;default:'[]'"                      json:"redirect_urls"`
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
