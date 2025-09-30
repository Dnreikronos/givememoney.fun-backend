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
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	Provider    string    `json:"provider"`
	ProviderID  string    `json:"provider_id"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Wallet      *Wallet   `json:"wallet,omitempty"`
}

type Wallet struct {
	ID       uuid.UUID `json:"id"`
	Provider string    `json:"provider"`
	Hash     string    `json:"hash"`
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
