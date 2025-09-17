package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

type TwitchProvider struct {
	handler *TwitchAuthHandler
}

func NewTwitchProvider() *TwitchProvider {
	return &TwitchProvider{
		handler: NewTwitchAuthHandler(),
	}
}

func (t *TwitchProvider) GetAuthURL() string {
	return t.handler.GenerateAuthURL()
}

func (t *TwitchProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	return t.handler.ExchangeCode(ctx, code)
}

func (t *TwitchProvider) GetUser(ctx context.Context, token string) (utils.ProviderUser, error) {
	user, err := t.handler.GetUser(ctx, token)
	if err != nil {
		return utils.ProviderUser{}, err
	}

	return utils.ProviderUser{
		ID:    user.ID,
		Name:  user.DisplayName,
		Email: user.Email,
	}, nil
}

func (t *TwitchProvider) GetProviderType() utils.StreamerProvider {
	return utils.ProviderTwitch
}
