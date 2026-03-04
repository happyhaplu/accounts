package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingCustomer maps a workspace to a Stripe customer.
// Created lazily on first checkout. One record per workspace.
type BillingCustomer struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"          json:"id"`
	WorkspaceID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"workspace_id"`
	Workspace        Workspace `gorm:"foreignKey:WorkspaceID"         json:"-"`
	StripeCustomerID string    `gorm:"not null;uniqueIndex"           json:"stripe_customer_id"`
	CreatedAt        time.Time `                                      json:"created_at"`
	UpdatedAt        time.Time `                                      json:"updated_at"`
}

func (b *BillingCustomer) BeforeCreate(_ *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
