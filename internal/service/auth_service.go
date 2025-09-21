package service

import (
	"context"
	"fmt"
	"log"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

type AuthService struct {
	streamerRepo interfaces.StreamerRepositoryInterface
	providers    map[utils.StreamerProvider]utils.AuthProvider
}

func NewAuthService(streamerRepo interfaces.StreamerRepositoryInterface) *AuthService {
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

func (s *AuthService) GetProvider(provider utils.StreamerProvider) (utils.AuthProvider, bool) {
	p, exists := s.providers[provider]
	return p, exists
}

func (s *AuthService) UpsertStreamer(ctx context.Context, provider utils.StreamerProvider, user utils.ProviderUser) (*model.Streamer, error) {
	return s.upsertStreamer(ctx, provider, user)
}

func (s *AuthService) upsertStreamer(ctx context.Context, provider utils.StreamerProvider, user utils.ProviderUser) (*model.Streamer, error) {
	existingStreamer, err := s.streamerRepo.FindByProvider(ctx, provider, user.ID)
	if err == nil {
		existingStreamer.Name = user.Name
		existingStreamer.Email = user.Email

		if err := s.streamerRepo.Update(ctx, existingStreamer); err != nil {
			log.Printf("Failed to update streamer: %v", err)
			return nil, fmt.Errorf("failed to update streamer: %w", err)
		}

		return existingStreamer, nil
	}

	tempHash := fmt.Sprintf("temp_%s_%s", provider, user.ID)
	wallet := &model.Wallet{
		WalletProvider:   utils.MetamaskWalletProvider,
		WalletProviderID: "",
		Hash:             tempHash,
	}

	streamer := &model.Streamer{
		Provider:   provider,
		ProviderID: user.ID,
		Name:       user.Name,
		Email:      user.Email,
	}

	if err := s.streamerRepo.CreateWithWallet(ctx, streamer, wallet); err != nil {
		log.Printf("Failed to create streamer with wallet: %v", err)
		return nil, fmt.Errorf("failed to create streamer: %w", err)
	}

	createdStreamer, err := s.streamerRepo.FindByID(ctx, streamer.ID)
	if err != nil {
		log.Printf("Failed to reload streamer: %v", err)
		return nil, fmt.Errorf("failed to reload streamer: %w", err)
	}

	return createdStreamer, nil
}
