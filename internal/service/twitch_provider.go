package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

type TwitchProvider struct {
	handler *TwitchAuthHandler
}

func NewTwitchProvider() utils.AuthProvider {
	return &TwitchProvider{
		handler: NewTwitchAuthHandler(),
	}
}

func (p *TwitchProvider) GetAuthURL() string {
	return p.handler.GenerateAuthURL()
}

func (p *TwitchProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	return p.handler.ExchangeCode(ctx, code)
}

func (p *TwitchProvider) GetUser(ctx context.Context, token string) (utils.ProviderUser, error) {
	twitchUser, err := p.handler.GetUser(ctx, token)
	if err != nil {
		return utils.ProviderUser{}, err
	}

	// Convert TwitchUser to ProviderUser
	return utils.ProviderUser{
		ID:    twitchUser.ID,
		Name:  twitchUser.DisplayName,
		Email: twitchUser.Email,
	}, nil
}

func (p *TwitchProvider) GetProviderType() utils.StreamerProvider {
	return utils.ProviderTwitch
}

func (p *TwitchProvider) ValidateState(state string) bool {
	return p.handler.ValidateState(state)
}