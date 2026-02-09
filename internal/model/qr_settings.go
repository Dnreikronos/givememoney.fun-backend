package model

import (
	"time"

	"github.com/google/uuid"
)

type QRSettingsRequest struct {
	PixelColor      string `json:"pixel_color" binding:"required"`
	BackgroundColor string `json:"background_color" binding:"required"`
	BorderColor     string `json:"border_color" binding:"required"`
	LogoUrl         string `json:"logo_url"`
	LogoSize        int    `json:"logo_size" binding:"min=10,max=30"`
	LogoShape       string `json:"logo_shape" binding:"required,oneof=circle square"`
	TopText         string `json:"top_text"`
	BottomText      string `json:"bottom_text"`
	TextColor       string `json:"text_color" binding:"required"`
	TextSize        string `json:"text_size" binding:"required,oneof=small medium large"`
	FrameStyle      string `json:"frame_style" binding:"required,oneof=none rounded square neon"`
	QRSize          int    `json:"qr_size" binding:"required,min=200,max=600"`
	PixelPattern    string `json:"pixel_pattern" binding:"required,oneof=square rounded dots"`
}

type QRSettings struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	StreamerID      uuid.UUID `json:"streamer_id" gorm:"type:uuid;not null;uniqueIndex"`
	Streamer        Streamer  `json:"-" gorm:"foreignKey:StreamerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PixelColor      string    `json:"pixel_color" gorm:"not null;default:'#000000'"`
	BackgroundColor string    `json:"background_color" gorm:"not null;default:'#ffffff'"`
	BorderColor     string    `json:"border_color" gorm:"not null;default:'#00a896'"`
	LogoUrl         string    `json:"logo_url" gorm:"default:null"`
	LogoSize        int       `json:"logo_size" gorm:"not null;default:20"`
	LogoShape       string    `json:"logo_shape" gorm:"not null;default:'circle'"`
	TopText         string    `json:"top_text" gorm:"default:''"`
	BottomText      string    `json:"bottom_text" gorm:"default:''"`
	TextColor       string    `json:"text_color" gorm:"not null;default:'#1f2235'"`
	TextSize        string    `json:"text_size" gorm:"not null;default:'medium'"`
	FrameStyle      string    `json:"frame_style" gorm:"not null;default:'rounded'"`
	QRSize          int       `json:"qr_size" gorm:"not null;default:300"`
	PixelPattern    string    `json:"pixel_pattern" gorm:"not null;default:'square'"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
