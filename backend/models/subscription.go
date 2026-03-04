package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Subscription records which product plan a workspace is on.
// status values: "active" | "canceled" | "expired"
type Subscription struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey"                                   json:"id"`
	WorkspaceID          uuid.UUID `gorm:"type:uuid;not null;index"                               json:"workspace_id"`
	Workspace            Workspace `gorm:"foreignKey:WorkspaceID"                                 json:"workspace,omitempty"`
	ProductID            uuid.UUID `gorm:"type:uuid;not null;index"                               json:"product_id"`
	Product              Product   `gorm:"foreignKey:ProductID"                                   json:"product,omitempty"`
	PlanName             string    `gorm:"type:varchar(80);not null"                              json:"plan_name"`
	Status               string    `gorm:"type:varchar(20);not null;default:'active';index"       json:"status"`
	CurrentPeriodEnd     time.Time `gorm:"not null"                                               json:"current_period_end"`
	StripeSubscriptionID *string   `gorm:"uniqueIndex"                                            json:"stripe_subscription_id,omitempty"`
	CreatedAt            time.Time `                                                               json:"created_at"`
	UpdatedAt            time.Time `                                                               json:"updated_at"`
}

func (s *Subscription) BeforeCreate(_ *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// IsAccessible returns true when the subscription is active and not yet expired.
func (s *Subscription) IsAccessible() bool {
	return s.Status == "active" && time.Now().Before(s.CurrentPeriodEnd)
}
