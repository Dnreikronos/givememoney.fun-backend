package dto

import (
	"time"

	"github.com/google/uuid"
)

// AuthResponse represents a successful authentication response
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents user information in auth responses
type UserInfo struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Username    string    `json:"username"`
	Provider    string    `json:"provider"`
	ProviderID  string    `json:"provider_id"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	Wallet      *Wallet   `json:"wallet,omitempty"`
}

// Wallet represents wallet information in responses
type Wallet struct {
	ID       uuid.UUID `json:"id"`
	Provider string    `json:"provider"`
	Hash     string    `json:"hash"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	Error   string                 `json:"error"`
	Fields  map[string]string      `json:"fields"`
	Details map[string]interface{} `json:"details,omitempty"`
}
