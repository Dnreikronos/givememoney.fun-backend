package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	StreamerID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"streamer_id"`
	SessionToken   string     `gorm:"type:varchar(255);not null;unique" json:"session_token"`
	RefreshTokenID *uuid.UUID `gorm:"type:uuid;index" json:"refresh_token_id,omitempty"`
	ExpiresAt      time.Time  `gorm:"not null" json:"expires_at"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastAccessedAt time.Time  `gorm:"autoCreateTime" json:"last_accessed_at"`
	UserAgent      string     `gorm:"type:text" json:"user_agent,omitempty"`
	IPAddress      string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	Country        string     `gorm:"type:varchar(2)" json:"country,omitempty"`
	DeviceType     string     `gorm:"type:varchar(50)" json:"device_type,omitempty"`
	LoginMethod    string     `gorm:"type:varchar(50)" json:"login_method,omitempty"`

	Streamer     Streamer      `gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"streamer,omitempty"`
	RefreshToken *RefreshToken `gorm:"foreignKey:RefreshTokenID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"refresh_token,omitempty"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return !s.IsExpired() && s.IsActive
}

func (s *Session) UpdateLastAccessed() {
	s.LastAccessedAt = time.Now()
}

func (s *Session) Deactivate() {
	s.IsActive = false
}

func (s *Session) ExtendExpiry(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}

type SessionInfo struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	UserAgent      string    `json:"user_agent,omitempty"`
	IPAddress      string    `json:"ip_address,omitempty"`
	Country        string    `json:"country,omitempty"`
	DeviceType     string    `json:"device_type,omitempty"`
	LoginMethod    string    `json:"login_method,omitempty"`
	IsActive       bool      `json:"is_active"`
	IsCurrent      bool      `json:"is_current"`
}

func (s *Session) ToSessionInfo(isCurrent bool) SessionInfo {
	return SessionInfo{
		ID:             s.ID,
		CreatedAt:      s.CreatedAt,
		LastAccessedAt: s.LastAccessedAt,
		ExpiresAt:      s.ExpiresAt,
		UserAgent:      s.UserAgent,
		IPAddress:      s.IPAddress,
		Country:        s.Country,
		DeviceType:     s.DeviceType,
		LoginMethod:    s.LoginMethod,
		IsActive:       s.IsActive,
		IsCurrent:      isCurrent,
	}
}
