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
			utils.ProviderKick:   NewKickProvider(),
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

func (s *AuthService) RegisterWithEmail(ctx context.Context, name, email, password string) (*model.Streamer, error) {
	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	streamer := &model.Streamer{
		Provider:     utils.ProviderTwitch, // Using a default provider for email users
		ProviderID:   email,                // Using email as ProviderID for email auth
		Name:         name,
		Email:        email,
		PasswordHash: &hashedPassword,
	}

	createdStreamer, err := s.streamerRepo.FindByID(ctx, streamer.ID)
	if err != nil {
		log.Printf("Failed to reload streamer: %v", err)
		return nil, fmt.Errorf("failed to reload streamer: %w", err)
	}

	return createdStreamer, nil
}

func (s *AuthService) AuthenticateWithEmail(ctx context.Context, email, password string) (*model.Streamer, error) {
	streamer, err := s.streamerRepo.FindByEmail(ctx, email)
	if err != nil {
		log.Printf("Failed to find user by email: %v", err)
		return nil, fmt.Errorf("invalid credentials")
	}

	if streamer.PasswordHash == nil {
		log.Printf("User %s has no password set", email)
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := utils.VerifyPassword(password, *streamer.PasswordHash); err != nil {
		log.Printf("Password verification failed for user %s: %v", email, err)
		return nil, fmt.Errorf("invalid credentials")
	}

	return streamer, nil
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

	streamer := &model.Streamer{
		Provider:   provider,
		ProviderID: user.ID,
		Name:       user.Name,
		Email:      user.Email,
	}

	if err := s.streamerRepo.CreateStreamer(ctx, streamer); err != nil {
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
