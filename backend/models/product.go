package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a single product offering in the Outcraftly registry.
// New products are added by inserting a row — no code changes required elsewhere.
type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"                          json:"id"`
	Name        string    `gorm:"type:varchar(80);uniqueIndex;not null"          json:"name"`
	Description string    `gorm:"type:text"                                     json:"description"`
	IsActive    bool      `gorm:"not null;default:true"                         json:"is_active"`
	CreatedAt   time.Time `                                                     json:"created_at"`
}

func (p *Product) BeforeCreate(_ *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
