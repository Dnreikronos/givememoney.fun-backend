package model

import (
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Streamer struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey"`
	Provider   utils.StreamerProvider `json:"provider" gorm:"not null"`
	ProviderID string                 `json:"provider_id" gorm:"not null; index:idx_provider_id"`
	Name       string                 `json:"name,omitempty"`
	Email      string                 `json:"email"`
	WalletID   uuid.UUID              `json:"wallet_id" gorm:"type:uuid;not null;index"`
	Wallet     Wallet                 `json:"wallet" gorm:"foreignKey:WalletID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt  time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

func (s *Streamer) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.New()
	return
}
