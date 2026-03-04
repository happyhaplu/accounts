package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WorkspaceInvite holds a pending invite sent to an email address.
// Status: "pending" | "accepted" | "revoked"
type WorkspaceInvite struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"                      json:"id"`
	WorkspaceID uuid.UUID  `gorm:"type:uuid;not null;index"                  json:"workspace_id"`
	Workspace   Workspace  `gorm:"foreignKey:WorkspaceID"                    json:"workspace,omitempty"`
	InvitedBy   uuid.UUID  `gorm:"type:uuid;not null"                        json:"invited_by"`
	InviterUser User       `gorm:"foreignKey:InvitedBy"                      json:"inviter,omitempty"`
	Email       string     `gorm:"not null;index"                            json:"email"`
	Role        string     `gorm:"type:varchar(20);not null;default:'member'" json:"role"`
	Token       string     `gorm:"uniqueIndex;not null"                      json:"-"`
	Status      string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (i *WorkspaceInvite) BeforeCreate(_ *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
