package model

import (
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

type Streamer struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider     utils.StreamerProvider `json:"provider" gorm:"not null"`
	ProviderID   string                 `json:"provider_id" gorm:"not null; index:idx_provider_id"`
	Name         string                 `json:"name,omitempty"`
	Email        string                 `json:"email"`
	PasswordHash *string                `json:"-" gorm:"type:varchar(255);default:null"`
	Wallet       *Wallet                `json:"wallet,omitempty" gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}
