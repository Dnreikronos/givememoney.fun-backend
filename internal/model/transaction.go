package model

import (
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

type TransactionRequest struct {
	Amount      float64        `json:"amount" validate:"required"`
	Message     string         `json:"message"`
	TxHash      string         `json:"tx_hash" validate:"required"`
	AddressFrom string         `json:"address_from" validate:"required"`
	Currency    utils.Currency `json:"currency" validate:"omitempty,oneof=ETH USDT USDC"`
}

type Transaction struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AddressFrom string         `json:"address_from" gorm:"type:string;not null"`
	AddressToID uuid.UUID      `json:"address_to_id" gorm:"type:uuid;not null;index"`
	AddressTo   Wallet         `json:"-" gorm:"foreignkey:AddressToID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`
	Amount      float64        `json:"amount" gorm:"type:decimal(20, 0);not null"`
	Currency    utils.Currency `json:"currency" gorm:"type:varchar(10);not null;default:'ETH'"`
	TxHash      string         `json:"tx_hash" gorm:"type:string;not null;unique;index"`
	Status      string         `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	Message     string         `json:"message" gorm:"type:string"`
}
