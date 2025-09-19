package service

import (
	"context"
	"fmt"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
)

type AuthService struct {
	streamerRepo *repository.StreamerRepository
	providers    map[utils.StreamerProvider]utils.AuthProvider
}

func NewAuthService(streamerRepo *repository.StreamerRepository) *AuthService {
	return &AuthService{
		streamerRepo: streamerRepo,
		providers: map[utils.StreamerProvider]utils.AuthProvider{
			utils.ProviderTwitch: NewTwitchProvider(),
		},
	}
}

func (s *AuthService) GetAuthURL(provider utils.StreamerProvider) (string, error) {
	p, exists := s.providers[provider]
	if !exists {
		return "", fmt.Errorf("unsupported provider")
	}
	return p.GetAuthURL(), nil
}

func (s *AuthService) Authenticate(ctx context.Context, provider utils.StreamerProvider, code string) (*model.Streamer, error) {
	p, exists := s.providers[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported provider")
	}

	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	user, err := p.GetUser(ctx, token)
	if err != nil {
		return nil, err
	}

	return s.upsertStreamer(ctx, provider, user)
}

func (s *AuthService) upsertStreamer(ctx context.Context, provider utils.StreamerProvider, user utils.ProviderUser) (*model.Streamer, error) {
	var streamer model.Streamer

	result := s.streamerRepo.GetDB().WithContext(ctx).Where("provider = ? AND provider_id = ?", provider, user.ID).First(&streamer)

	if result.Error != nil {

		streamer = model.Streamer{
			Provider:   provider,
			ProviderID: user.ID,
			Name:       user.Name,
			Email:      user.Email,
			WalletID:   uuid.New(),
		}

		if err := s.streamerRepo.GetDB().WithContext(ctx).Create(&streamer).Error; err != nil {
			return nil, err
		}
	} else {
		// Update existing
		streamer.Name = user.Name
		streamer.Email = user.Email

		if err := s.streamerRepo.GetDB().WithContext(ctx).Save(&streamer).Error; err != nil {
			return nil, err
		}
	}

	return &streamer, nil
}
