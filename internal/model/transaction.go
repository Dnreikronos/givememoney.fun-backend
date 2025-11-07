package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	AddressFrom string    `json:"address_from" gorm:"type:string;not null"`
	AddressToID uuid.UUID `json:"address_to_id" gorm:"type:uuid;not null;index"`
	AddressTo   Wallet    `json:"address_to" gorm:"foreignkey:AddressToID;constraint:OnUpdate:CASCADE;OnDelete:CASCADE"`
	Amount      float64   `json:"amount" gorm:"type:decimal(20, 0);not null"`
	TxHash      string    `json:"tx_hash" gorm:"type:string;not null;unique;index"`
	Status      string    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Message     string    `json:"message" gorm:"type:string"`
}
