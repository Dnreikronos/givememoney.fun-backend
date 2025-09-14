package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Streamer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email"`
	WalletID  uuid.UUID `json:"wallet_id" gorm:"type:uuid;not null;index"`
	Wallet    Wallet    `json:"wallet" gorm:"foreignKey:WalletID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

