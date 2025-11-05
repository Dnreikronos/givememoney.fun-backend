package model

import (
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

type WalletUpdateInput struct {
	Hash string `json:"hash" validate:"required,len=64"`
}

type WalletRequest struct {
	WalletProvider utils.WalletProvider `json:"wallet_provider" validate:"required"`
	Hash           string               `json:"hash" validate:"required,len=64"`
}

type Wallet struct {
	ID             uuid.UUID            `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WalletProvider utils.WalletProvider `json:"wallet_provider" gorm:"not null;type:varchar(50)"`
	WalletAddress  string               `json:"wallet_address" gorm:"not null;size:64;unique"`
	StreamerID     uuid.UUID            `json:"streamer_id" gorm:"type:uuid;not null;index"`
	Streamer       Streamer             `json:"streamer" gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt      time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
}
