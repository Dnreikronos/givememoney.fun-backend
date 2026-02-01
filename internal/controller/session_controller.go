package controller

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionController struct {
	sessionService *service.SessionService
	jwtService     *service.JWTService
	streamerRepo   interfaces.StreamerRepositoryInterface
	logger         *zap.Logger
	helpers        *AuthHelpers
}

// NewSessionController creates a new SessionController with the given dependencies.
func NewSessionController(
	sessionService *service.SessionService,
	jwtService *service.JWTService,
	streamerRepo interfaces.StreamerRepositoryInterface,
	helpers *AuthHelpers,
	logger *zap.Logger,
) *SessionController {
	return &SessionController{
		sessionService: sessionService,
		jwtService:     jwtService,
		streamerRepo:   streamerRepo,
		helpers:        helpers,
		logger:         logger,
	}
}

// Create creates a session from OAuth callback tokens and user info.
func (c *SessionController) Create(ctx *gin.Context) {
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
		DeviceType:  c.helpers.DetectDeviceType(ctx.GetHeader("User-Agent")),
		Country:     c.helpers.DetectCountry(ctx.ClientIP()),
	}

	_, tokenPair, err := c.sessionService.CreateSession(sessionReq)
	if err != nil {
		c.logger.Error("Session creation failed", zap.Error(err))
		appErr := errors.NewInternalError("Session creation failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.jwtService.SetTokenCookies(ctx, tokenPair)

	userResponse := c.helpers.GetUserResponse(streamer)
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    userResponse,
	})
}

// Get returns the current session and user if authenticated.
func (c *SessionController) Get(ctx *gin.Context) {
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

	userResponse := c.helpers.GetUserResponse(streamer)
	ctx.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"sessionId":     result.Session.ID,
		"hasToken":      true,
		"user":          userResponse,
	})
}

// Delete invalidates the current session.
func (c *SessionController) Delete(ctx *gin.Context) {
	err := c.sessionService.InvalidateSession(ctx)
	if err != nil {
		c.logger.Error("Session deletion failed", zap.Error(err))
		appErr := errors.NewInternalError("Session deletion failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Refresh refreshes the access token using the refresh token.
func (c *SessionController) Refresh(ctx *gin.Context) {
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

// GetActive returns the list of active sessions for the current user.
func (c *SessionController) GetActive(ctx *gin.Context) {
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

// Logout invalidates the current session (alias for Delete).
func (c *SessionController) Logout(ctx *gin.Context) {
	err := c.sessionService.InvalidateSession(ctx)
	if err != nil {
		c.logger.Error("Logout failed", zap.Error(err))
		appErr := errors.NewInternalError("Logout failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
