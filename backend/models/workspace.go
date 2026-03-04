package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Workspace is the tenant / subscription anchor created for every user on sign-up.
type Workspace struct {
	ID        uuid.UUID         `gorm:"type:uuid;primaryKey"           json:"id"`
	Name      string            `gorm:"not null"                        json:"name"`
	OwnerID   uuid.UUID         `gorm:"type:uuid;not null;index"       json:"owner_id"`
	Members   []WorkspaceMember `gorm:"foreignKey:WorkspaceID"         json:"members,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (w *Workspace) BeforeCreate(_ *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// WorkspaceMember links a User to a Workspace.
// Roles are exactly: "owner" | "member"
type WorkspaceMember struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"                            json:"id"`
	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_ws_user"     json:"workspace_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_ws_user"     json:"user_id"`
	User        User      `gorm:"foreignKey:UserID"                               json:"user"`
	Role        string    `gorm:"type:varchar(20);not null;default:'member'"     json:"role"`
	JoinedAt    time.Time `gorm:"autoCreateTime"                                  json:"joined_at"`
}

func (m *WorkspaceMember) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
