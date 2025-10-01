package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	StreamerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"streamer_id"`
	TokenHash  string     `gorm:"type:varchar(255);not null;unique" json:"token_hash"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	IsRevoked  bool       `gorm:"default:false" json:"is_revoked"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	UserAgent  string     `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress  string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`

	Streamer Streamer `gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"streamer,omitempty"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}

func (rt *RefreshToken) MarkAsUsed() {
	now := time.Now()
	rt.LastUsedAt = &now
}

func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
}
