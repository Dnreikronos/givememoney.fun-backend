package model

import (
	"time"

	"github.com/google/uuid"
)

type AlertSettingsRequest struct {
	BackgroundColor string `json:"background_color" binding:"required"`
	TextColor       string `json:"text_color" binding:"required"`
	MessageColor    string `json:"message_color" binding:"required"`
	AccentColor     string `json:"accent_color" binding:"required"`
	HeaderText      string `json:"header_text" binding:"required"`
	ShowDonorName   bool   `json:"show_donor_name"`
	ShowAmount      bool   `json:"show_amount"`
	ShowMessage     bool   `json:"show_message"`
	MinDuration     int    `json:"min_duration" binding:"required,min=2000,max=10000"`
	MaxDuration     int    `json:"max_duration" binding:"required,min=2000,max=10000"`
	SoundEnabled    bool   `json:"sound_enabled"`
	Position        string `json:"position" binding:"omitempty,oneof=top center bottom"`
}

type AlertSettings struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StreamerID      uuid.UUID `json:"streamer_id" gorm:"type:uuid;not null;uniqueIndex"`
	Streamer        Streamer  `json:"-" gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	BackgroundColor string    `json:"background_color" gorm:"not null;default:'#1f2235'"`
	TextColor       string    `json:"text_color" gorm:"not null;default:'#ffffff'"`
	MessageColor    string    `json:"message_color" gorm:"not null;default:'#cbd2dd'"`
	AccentColor     string    `json:"accent_color" gorm:"not null;default:'#00a896'"`
	HeaderText      string    `json:"header_text" gorm:"not null;default:'New Donation!'"`
	ShowDonorName   bool      `json:"show_donor_name" gorm:"not null;default:true"`
	ShowAmount      bool      `json:"show_amount" gorm:"not null;default:true"`
	ShowMessage     bool      `json:"show_message" gorm:"not null;default:true"`
	MinDuration     int       `json:"min_duration" gorm:"not null;default:3000"`
	MaxDuration     int       `json:"max_duration" gorm:"not null;default:8000"`
	SoundEnabled    bool      `json:"sound_enabled" gorm:"not null;default:true"`
	Position        string    `json:"position" gorm:"not null;default:'top'"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
