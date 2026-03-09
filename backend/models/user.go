package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User is the core identity record.
type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey"  json:"id"`
	Email               string     `gorm:"uniqueIndex;not null"  json:"email"`
	PasswordHash        string     `gorm:"not null"              json:"-"`

// Email verification (legacy link fields kept for DB compat, use OTP fields below)
        EmailVerified       bool       `gorm:"default:false"         json:"email_verified"`
        EmailVerifyToken    *string    `gorm:"index"                 json:"-"`
        EmailVerifyExpires  *time.Time `                              json:"-"`

        // OTP — shared for email verification and password reset
        OTPCode    *string    `gorm:"index" json:"-"`
        OTPExpires *time.Time `              json:"-"`
        OTPPurpose *string    `              json:"-"` // "email_verify" | "password_reset"

	// Profile
	Name             string `gorm:"default:''"`
	CompanyName      string `gorm:"default:''"`
	JobTitle         string `gorm:"default:''"`
	PhoneCountryCode string `gorm:"default:''"`
	PhoneNumber      string `gorm:"default:''"`
	ProfileComplete  bool   `gorm:"default:false" json:"profile_complete"`

	// Security
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	FailedLoginAttempts int        `gorm:"default:0" json:"-"`
	LockedUntil         *time.Time `json:"-"`

	// Password reset
	ResetToken        *string    `gorm:"index" json:"-"`
	ResetTokenExpires *time.Time `json:"-"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

// BeforeCreate auto-generates a UUID if one hasn't been set.
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
