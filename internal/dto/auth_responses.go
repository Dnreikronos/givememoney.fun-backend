package dto

import (
	"github.com/google/uuid"
)

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Wallets    []Wallet  `json:"wallets,omitempty"`
}

type Wallet struct {
	ID            uuid.UUID `json:"id"`
	Provider      string    `json:"provider"`
	WalletAddress string    `json:"wallet_address"`
}
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

type ValidationErrorResponse struct {
	Error   string                 `json:"error"`
	Fields  map[string]string      `json:"fields"`
	Details map[string]interface{} `json:"details,omitempty"`
}
