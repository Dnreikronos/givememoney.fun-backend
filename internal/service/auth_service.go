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
