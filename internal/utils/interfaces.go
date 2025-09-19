package utils

import "context"

type AuthProvider interface {
	GetAuthURL() string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUser(ctx context.Context, token string) (ProviderUser, error)
	GetProviderType() StreamerProvider
}
