package dto

import (
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

type TwitchCallbackRequest struct {
	Code  string `form:"code" validate:"required" binding:"required"`
	State string `form:"state" validate:"omitempty"`
}

type KickCallbackRequest struct {
	Code  string `form:"code" validate:"required" binding:"required"`
	State string `form:"state" validate:"required" binding:"required"`
}

type WalletRequest struct {
	WalletProvider utils.WalletProvider `json:"wallet_provider" validate:"required" binding:"required"`
	WalletAddress  string               `json:"wallet_address" validate:"required,len=64" binding:"required,len=64"`
}

type TransactionRequest struct {
	Amount      float64        `json:"amount" validate:"required"`
	Message     string         `json:"message"`
	TxHash      string         `json:"tx_hash" validate:"required"`
	AddressFrom string         `json:"address_from" validate:"required"`
	Currency    utils.Currency `json:"currency" validate:"omitempty,oneof=ETH SOL USDT USDC"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" binding:"required"`
}

type SessionLogoutRequest struct {
	LogoutAll bool `json:"logout_all" validate:"omitempty"`
}

type RegisterRequest struct {
	Name            string `json:"name" validate:"required,min=2,max=50" binding:"required,min=2,max=50"`
	Email           string `json:"email" validate:"required,email" binding:"required,email"`
	Password        string `json:"password" validate:"required,min=8" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required" binding:"required"`
}

type EmailLoginRequest struct {
	Email    string `json:"email" validate:"required,email" binding:"required,email"`
	Password string `json:"password" validate:"required" binding:"required"`
}
