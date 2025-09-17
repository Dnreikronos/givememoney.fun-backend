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
