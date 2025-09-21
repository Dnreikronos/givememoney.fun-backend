package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthController struct {
	authService      *service.AuthService
	jwtService       *service.JWTService
	sessionService   *service.SessionService
	streamerRepo     interfaces.StreamerRepositoryInterface
	logger           *zap.Logger
}

func NewAuthController(
	authService *service.AuthService,
	jwtService *service.JWTService,
	sessionService *service.SessionService,
	streamerRepo interfaces.StreamerRepositoryInterface,
	logger *zap.Logger,
) *AuthController {
	return &AuthController{
		authService:      authService,
		jwtService:       jwtService,
		sessionService:   sessionService,
		streamerRepo:     streamerRepo,
		logger:           logger,
	}
}

func (c *AuthController) TwitchLogin(ctx *gin.Context) {
	authURL, err := c.authService.GetAuthURL(utils.ProviderTwitch)
	if err != nil {
		appErr := errors.NewTwitchAPIError("Failed to get Twitch auth URL", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.logger.Info("Redirecting to Twitch auth",
		zap.String("auth_url", authURL),
		zap.String("ip", ctx.ClientIP()),
	)

	ctx.Redirect(http.StatusFound, authURL)
}

func (c *AuthController) TwitchCallback(ctx *gin.Context) {
	var req dto.TwitchCallbackRequest
	if !middleware.ValidateQuery(ctx, &req) {
		return
	}

	frontendURL := c.getFrontendURL()
	c.logger.Info("Processing Twitch callback",
		zap.String("frontend_url", frontendURL),
		zap.String("ip", ctx.ClientIP()),
	)

	streamer, err := c.authService.Authenticate(ctx.Request.Context(), utils.ProviderTwitch, req.Code)
	if err != nil {
		c.logger.Error("Twitch authentication failed", zap.Error(err))
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=auth_failed", frontendURL))
		return
	}

	tokenPair, err := c.createSessionForStreamer(ctx, streamer)
	if err != nil {
		c.logger.Error("Session creation failed", zap.Error(err), zap.String("streamer_id", streamer.ID.String()))
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=session_creation_failed", frontendURL))
		return
	}

	c.jwtService.SetTokenCookies(ctx, tokenPair)

	c.logger.Info("Session created successfully",
		zap.String("streamer_id", streamer.ID.String()),
		zap.String("provider", string(streamer.Provider)),
	)
	ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/dashboard", frontendURL))
}

func (c *AuthController) TwitchToken(ctx *gin.Context) {
	code := ctx.PostForm("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing authorization code"})
		return
	}

	provider, exists := c.authService.GetProvider(utils.ProviderTwitch)
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "provider not found"})
		return
	}

	accessToken, err := provider.ExchangeCode(ctx.Request.Context(), code)
	if err != nil {
			c.logger.Error("Token exchange failed", zap.Error(err))
		appErr := errors.NewTwitchAPIError("Token exchange failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	twitchUser, err := provider.GetUser(ctx.Request.Context(), accessToken)
	if err != nil {
			c.logger.Error("Failed to get user info", zap.Error(err))
		appErr := errors.NewTwitchAPIError("Failed to get user info", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	streamer, err := c.authService.UpsertStreamer(ctx.Request.Context(), utils.ProviderTwitch, twitchUser)
	if err != nil {
			c.logger.Error("Failed to upsert streamer", zap.Error(err))
		appErr := errors.NewDatabaseError("upsert streamer", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	jwt, err := c.jwtService.GenerateToken(streamer.ID, streamer.Email, streamer.Name, string(streamer.Provider))
	if err != nil {
			c.logger.Error("Failed to generate JWT", zap.Error(err))
		appErr := errors.NewInternalError("Failed to generate token", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 1 hour
		"scope":         "user:read:email",
		"refresh_token": "",
		"jwt_token": jwt,
	})
}

func (c *AuthController) TwitchUser(ctx *gin.Context) {
	token, err := utils.ExtractBearerToken(ctx.GetHeader("Authorization"))
	if err != nil {
		appErr := errors.NewUnauthorizedError("Invalid authorization header")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	if claims, err := c.jwtService.ValidateToken(token); err == nil {
		streamer, err := c.getStreamerByID(ctx, claims.UserID)
		if err != nil {
			appErr := errors.NewNotFoundError("user")
			middleware.AbortWithError(ctx, appErr)
			return
		}

		userResponse := c.getUserResponse(streamer)
		ctx.JSON(http.StatusOK, userResponse)
		return
	}

	streamer, err := c.validateTwitchToken(ctx, token)
	if err != nil {
		appErr := errors.NewUnauthorizedError("Invalid token")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	userResponse := c.getUserResponse(streamer)
	ctx.JSON(http.StatusOK, userResponse)
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	tokenPair, err := c.sessionService.RefreshSession(ctx)
	if err != nil {
			c.logger.Error("Token refresh failed", zap.Error(err))
		appErr := errors.NewUnauthorizedError("Invalid refresh token")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":    true,
		"expires_in": tokenPair.ExpiresIn,
	})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	err := c.sessionService.InvalidateSession(ctx)
	if err != nil {
			c.logger.Error("Logout failed", zap.Error(err))
		appErr := errors.NewInternalError("Logout failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}


func (c *AuthController) CreateSession(ctx *gin.Context) {
	var request struct {
		AccessToken  string `json:"access_token" binding:"required"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			ID       string `json:"id" binding:"required"`
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Provider string `json:"provider" binding:"required"`
		} `json:"user" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	userID, err := uuid.Parse(request.User.ID)
	if err != nil {
		appErr := errors.NewValidationError("Invalid user ID format", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	streamer, err := c.streamerRepo.FindByID(ctx.Request.Context(), userID)
	if err != nil {
		appErr := errors.NewNotFoundError("user")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	sessionReq := service.SessionCreateRequest{
		StreamerID:  streamer.ID,
		UserAgent:   ctx.GetHeader("User-Agent"),
		IPAddress:   ctx.ClientIP(),
		LoginMethod: request.User.Provider,
		DeviceType:  c.detectDeviceType(ctx.GetHeader("User-Agent")),
		Country:     c.detectCountry(ctx.ClientIP()),
	}

	_, tokenPair, err := c.sessionService.CreateSession(sessionReq)
	if err != nil {
			c.logger.Error("Session creation failed", zap.Error(err))
		appErr := errors.NewInternalError("Session creation failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.jwtService.SetTokenCookies(ctx, tokenPair)

	userResponse := c.getUserResponse(streamer)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    userResponse,
	})
}

func (c *AuthController) GetSession(ctx *gin.Context) {
	result, err := c.sessionService.ValidateSession(ctx)
	if err != nil || !result.IsValid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	streamer, err := c.streamerRepo.FindByID(ctx.Request.Context(), result.Claims.UserID)
	if err != nil {
		appErr := errors.NewNotFoundError("user")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	userResponse := c.getUserResponse(streamer)
	ctx.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"sessionId":     result.Session.ID,
		"hasToken":      true,
		"user":          userResponse,
	})
}

func (c *AuthController) DeleteSession(ctx *gin.Context) {
	err := c.sessionService.InvalidateSession(ctx)
	if err != nil {
			c.logger.Error("Session deletion failed", zap.Error(err))
		appErr := errors.NewInternalError("Session deletion failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func (c *AuthController) GetActiveSessions(ctx *gin.Context) {
	result, err := c.sessionService.ValidateSession(ctx)
	if err != nil || !result.IsValid {
			appErr := errors.NewUnauthorizedError("Authentication required")
		middleware.AbortWithError(ctx, appErr)
		return
	}

	sessions, err := c.sessionService.GetActiveSessions(result.Claims.UserID, &result.Session.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sessions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sessions": sessions})
}


func (c *AuthController) getFrontendURL() string {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	return frontendURL
}

func (c *AuthController) createSessionForStreamer(ctx *gin.Context, streamer *model.Streamer) (*service.TokenPair, error) {
	sessionReq := service.SessionCreateRequest{
		StreamerID:  streamer.ID,
		UserAgent:   ctx.GetHeader("User-Agent"),
		IPAddress:   ctx.ClientIP(),
		LoginMethod: string(streamer.Provider),
		DeviceType:  c.detectDeviceType(ctx.GetHeader("User-Agent")),
		Country:     c.detectCountry(ctx.ClientIP()),
	}

	_, tokenPair, err := c.sessionService.CreateSession(sessionReq)
	return tokenPair, err
}

func (c *AuthController) getUserResponse(streamer *model.Streamer) dto.UserInfo {
	response := dto.UserInfo{
		ID:          streamer.ID,
		DisplayName: streamer.Name,
		Username:    streamer.Name,
		Provider:    string(streamer.Provider),
		ProviderID:  streamer.ProviderID,
	}

	if streamer.Wallet.ID != uuid.Nil {
		response.Wallet = &dto.Wallet{
			ID:       streamer.Wallet.ID,
			Provider: string(streamer.Wallet.WalletProvider),
			Hash:     streamer.Wallet.Hash,
		}
	}

	return response
}

func (c *AuthController) getStreamerByID(ctx *gin.Context, streamerID uuid.UUID) (*model.Streamer, error) {
	return c.streamerRepo.FindByID(ctx.Request.Context(), streamerID)
}

func (c *AuthController) validateTwitchToken(ctx *gin.Context, token string) (*model.Streamer, error) {
	provider, exists := c.authService.GetProvider(utils.ProviderTwitch)
	if !exists {
		return nil, errors.NewInternalError("Twitch provider not found", nil)
	}

	twitchUser, err := provider.GetUser(ctx.Request.Context(), token)
	if err != nil {
		return nil, errors.NewTwitchAPIError("Failed to validate Twitch token", err)
	}

	streamer, err := c.streamerRepo.FindByProvider(ctx.Request.Context(), utils.ProviderTwitch, twitchUser.ID)
	if err != nil {
		return nil, errors.NewNotFoundError("user")
	}

	return streamer, nil
}

func (c *AuthController) detectDeviceType(userAgent string) string {
	if userAgent == "" {
		return "unknown"
	}

	userAgent = fmt.Sprintf("%s", userAgent)

	if contains(userAgent, "Mobile") || contains(userAgent, "Android") || contains(userAgent, "iPhone") {
		return "mobile"
	}
	if contains(userAgent, "Tablet") || contains(userAgent, "iPad") {
		return "tablet"
	}
	return "desktop"
}

func (c *AuthController) detectCountry(ipAddress string) string {
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
