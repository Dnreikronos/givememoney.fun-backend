package dto

import "github.com/Dnreikronos/givememoney.fun-backend/internal/utils"

// TwitchCallbackRequest represents the callback request from Twitch OAuth
type TwitchCallbackRequest struct {
	Code  string `form:"code" validate:"required" binding:"required"`
	State string `form:"state" validate:"omitempty"`
}

// KickCallbackRequest represents the callback request from Kick OAuth
type KickCallbackRequest struct {
	Code  string `form:"code" validate:"required" binding:"required"`
	State string `form:"state" validate:"required" binding:"required"`
}

// WalletRequest represents a wallet creation/update request
type WalletRequest struct {
	WalletProvider utils.WalletProvider `json:"wallet_provider" validate:"required" binding:"required"`
	Hash           string               `json:"hash" validate:"required,len=64" binding:"required,len=64"`
}

// RefreshTokenRequest represents a token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" binding:"required"`
}

// SessionLogoutRequest represents a logout request
type SessionLogoutRequest struct {
	LogoutAll bool `json:"logout_all" validate:"omitempty"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=50" binding:"required,min=2,max=50"`
	Email           string `json:"email" validate:"required,email" binding:"required,email"`
	Password        string `json:"password" validate:"required,min=8" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required" binding:"required"`
}

// EmailLoginRequest represents an email/password login request
type EmailLoginRequest struct {
	Email    string `json:"email" validate:"required,email" binding:"required,email"`
	Password string `json:"password" validate:"required" binding:"required"`
}
