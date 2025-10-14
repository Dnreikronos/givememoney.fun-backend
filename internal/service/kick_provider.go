package service

import (
	"context"
	"fmt"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

type KickProvider struct {
	handler *KickAuthHandler
}

func NewKickProvider() utils.AuthProvider {
	return &KickProvider{
		handler: NewKickAuthHandler(),
	}
}

func (p *KickProvider) GetAuthURL() string {
	return p.handler.GenerateAuthURL()
}

func (p *KickProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	return "", fmt.Errorf("state parameter required for Kick PKCE flow")
}

func (p *KickProvider) GetUser(ctx context.Context, token string) (utils.ProviderUser, error) {
	kickUser, err := p.handler.GetUser(ctx, token)
	if err != nil {
		return utils.ProviderUser{}, err
	}

	// Convert KickUser to ProviderUser
	return utils.ProviderUser{
		ID:    kickUser.ID,
		Name:  kickUser.DisplayName,
		Email: kickUser.Email,
	}, nil
}

func (p *KickProvider) GetProviderType() utils.StreamerProvider {
	return utils.ProviderKick
}

func (p *KickProvider) ValidateState(state string) bool {
	return p.handler.ValidateState(state)
}

// ExchangeCodeWithState provides state validation for PKCE flow
func (p *KickProvider) ExchangeCodeWithState(ctx context.Context, code string, state string) (string, error) {
	return p.handler.ExchangeCode(ctx, code, state)
}
