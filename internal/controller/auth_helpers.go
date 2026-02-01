package controller

import (
	"os"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthHelpers provides shared auth/session helpers for controllers.
type AuthHelpers struct {
	sessionService *service.SessionService
	streamerRepo   interfaces.StreamerRepositoryInterface
}

// NewAuthHelpers creates a new AuthHelpers with the given dependencies.
func NewAuthHelpers(
	sessionService *service.SessionService,
	streamerRepo interfaces.StreamerRepositoryInterface,
) *AuthHelpers {
	return &AuthHelpers{
		sessionService: sessionService,
		streamerRepo:   streamerRepo,
	}
}

func (h *AuthHelpers) GetFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return frontendURL
}

func (h *AuthHelpers) CreateSessionForStreamer(ctx *gin.Context, streamer *model.Streamer) (*service.TokenPair, error) {
	sessionReq := service.SessionCreateRequest{
		StreamerID:  streamer.ID,
		UserAgent:   ctx.GetHeader("User-Agent"),
		IPAddress:   ctx.ClientIP(),
		LoginMethod: string(streamer.Provider),
		DeviceType:  h.DetectDeviceType(ctx.GetHeader("User-Agent")),
		Country:     h.DetectCountry(ctx.ClientIP()),
	}

	_, tokenPair, err := h.sessionService.CreateSession(sessionReq)
	return tokenPair, err
}

func (h *AuthHelpers) GetUserResponse(streamer *model.Streamer) dto.UserInfo {
	provider := string(streamer.Provider)
	providerID := streamer.ProviderID

	if streamer.Provider == utils.ProviderEmail {
		provider = string(utils.ProviderEmail)
		providerID = streamer.Email
	}

	return dto.UserInfo{
		ID:         streamer.ID,
		Name:       streamer.Name,
		Email:      streamer.Email,
		Provider:   provider,
		ProviderID: providerID,
	}
}

func (h *AuthHelpers) GetStreamerByID(ctx *gin.Context, streamerID uuid.UUID) (*model.Streamer, error) {
	return h.streamerRepo.FindByID(ctx.Request.Context(), streamerID)
}

func (h *AuthHelpers) ValidateTwitchToken(ctx *gin.Context, token string, authService *service.AuthService) (*model.Streamer, error) {
	provider, exists := authService.GetProvider(utils.ProviderTwitch)
	if !exists {
		return nil, errors.NewInternalError("Twitch provider not found", nil)
	}

	twitchUser, err := provider.GetUser(ctx.Request.Context(), token)
	if err != nil {
		return nil, errors.NewProviderAPIError("Failed to validate Twitch token", err)
	}

	streamer, err := h.streamerRepo.FindByProvider(ctx.Request.Context(), utils.ProviderTwitch, twitchUser.ID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return streamer, nil
}

func (h *AuthHelpers) DetectDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	if contains(userAgent, "Mobile") || contains(userAgent, "Android") || contains(userAgent, "iPhone") {
		return "mobile"
	}
	if contains(userAgent, "Tablet") || contains(userAgent, "iPad") {
		return "tablet"
	}
	return "desktop"
}

func (h *AuthHelpers) DetectCountry(ipAddress string) string {
	if ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "XX"
	}
	return "XX"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
