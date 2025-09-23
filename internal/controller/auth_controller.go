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
	authService    *service.AuthService
	jwtService     *service.JWTService
	sessionService *service.SessionService
	streamerRepo   interfaces.StreamerRepositoryInterface
	logger         *zap.Logger
}

func NewAuthController(
	authService *service.AuthService,
	jwtService *service.JWTService,
	sessionService *service.SessionService,
	streamerRepo interfaces.StreamerRepositoryInterface,
	logger *zap.Logger,
) *AuthController {
	return &AuthController{
		authService:    authService,
		jwtService:     jwtService,
		sessionService: sessionService,
		streamerRepo:   streamerRepo,
		logger:         logger,
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

func (c *AuthController) KickLogin(ctx *gin.Context) {
	authURL, err := c.authService.GetAuthURL(utils.ProviderKick)
	if err != nil {
		appErr := errors.NewTwitchAPIError("Failed to get Kick auth URL", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.logger.Info("Redirecting to Kick auth",
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

func (c *AuthController) KickCallback(ctx *gin.Context) {
	var req dto.KickCallbackRequest
	if !middleware.ValidateQuery(ctx, &req) {
		return
	}

	frontendURL := c.getFrontendURL()
	c.logger.Info("Processing Kick callback",
		zap.String("frontend_url", frontendURL),
		zap.String("ip", ctx.ClientIP()),
	)

	// Validate state for PKCE flow
	provider, exists := c.authService.GetProvider(utils.ProviderKick)
	if !exists {
		c.logger.Error("Kick provider not found")
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=provider_not_found", frontendURL))
		return
	}

	// Type assert to access state validation (specific to our Kick provider)
	if kickProvider, ok := provider.(*service.KickProvider); ok {
		if !kickProvider.ValidateState(req.State) {
			c.logger.Error("Invalid state parameter", zap.String("state", req.State))
			ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=invalid_state", frontendURL))
			return
		}

		// Use ExchangeCodeWithState for proper PKCE flow
		token, err := kickProvider.ExchangeCodeWithState(ctx.Request.Context(), req.Code, req.State)
		if err != nil {
			c.logger.Error("Kick token exchange failed", zap.Error(err))
			ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=auth_failed", frontendURL))
			return
		}

		user, err := provider.GetUser(ctx.Request.Context(), token)
		if err != nil {
			c.logger.Error("Failed to get Kick user", zap.Error(err))
			ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=user_fetch_failed", frontendURL))
			return
		}

		streamer, err := c.authService.UpsertStreamer(ctx.Request.Context(), utils.ProviderKick, user)
		if err != nil {
			c.logger.Error("Failed to upsert Kick streamer", zap.Error(err))
			ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=user_creation_failed", frontendURL))
			return
		}

		tokenPair, err := c.createSessionForStreamer(ctx, streamer)
		if err != nil {
			c.logger.Error("Session creation failed", zap.Error(err), zap.String("streamer_id", streamer.ID.String()))
			ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=session_creation_failed", frontendURL))
			return
		}

		c.jwtService.SetTokenCookies(ctx, tokenPair)

		c.logger.Info("Kick session created successfully",
			zap.String("streamer_id", streamer.ID.String()),
			zap.String("provider", string(streamer.Provider)),
		)
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/dashboard", frontendURL))
		return
	}

	c.logger.Error("Failed to cast provider to KickProvider")
	ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=auth_failed", frontendURL))
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
		"jwt_token":     jwt,
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

func (c *AuthController) EmailRegister(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	// Validate password and confirmation using middleware helper
	if !middleware.ValidateRegisterRequest(ctx, &req) {
		return
	}

	c.logger.Info("Processing email registration",
		zap.String("email", req.Email),
		zap.String("name", req.Name),
		zap.String("ip", ctx.ClientIP()),
	)

	// Check if user already exists
	existingStreamer, err := c.streamerRepo.FindByEmail(ctx.Request.Context(), req.Email)
	if err == nil && existingStreamer != nil {
		c.logger.Warn("Registration attempt with existing email",
			zap.String("email", req.Email),
			zap.String("ip", ctx.ClientIP()),
		)
		ctx.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "Email already registered",
			Message: "An account with this email already exists",
		})
		return
	}

	// Create user via AuthService
	streamer, err := c.authService.RegisterWithEmail(ctx.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.logger.Error("Email registration failed", zap.Error(err))
		appErr := errors.NewInternalError("Registration failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	// Create session for the new user
	tokenPair, err := c.createSessionForStreamer(ctx, streamer)
	if err != nil {
		c.logger.Error("Session creation failed after registration", zap.Error(err), zap.String("streamer_id", streamer.ID.String()))
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Registration successful but session creation failed",
			Message: "Please try logging in",
		})
		return
	}

	c.jwtService.SetTokenCookies(ctx, tokenPair)

	c.logger.Info("Email registration successful",
		zap.String("streamer_id", streamer.ID.String()),
		zap.String("email", req.Email),
	)

	userResponse := c.getUserResponse(streamer)
	ctx.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         userResponse,
	})
}

func (c *AuthController) EmailLogin(ctx *gin.Context) {
	var req dto.EmailLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.logger.Info("Processing email login",
		zap.String("email", req.Email),
		zap.String("ip", ctx.ClientIP()),
	)

	// Authenticate user via AuthService
	streamer, err := c.authService.AuthenticateWithEmail(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.logger.Warn("Email login failed",
			zap.String("email", req.Email),
			zap.String("ip", ctx.ClientIP()),
			zap.Error(err),
		)
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "Invalid credentials",
			Message: "Email or password is incorrect",
		})
		return
	}

	// Create session for the authenticated user
	tokenPair, err := c.createSessionForStreamer(ctx, streamer)
	if err != nil {
		c.logger.Error("Session creation failed after login", zap.Error(err), zap.String("streamer_id", streamer.ID.String()))
		appErr := errors.NewInternalError("Login failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.jwtService.SetTokenCookies(ctx, tokenPair)

	c.logger.Info("Email login successful",
		zap.String("streamer_id", streamer.ID.String()),
		zap.String("email", req.Email),
	)

	userResponse := c.getUserResponse(streamer)
	ctx.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         userResponse,
	})
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
	provider := string(streamer.Provider)
	providerID := streamer.ProviderID

	// For email-authenticated users, set provider to "email" for clarity
	if streamer.PasswordHash != nil && streamer.ProviderID == streamer.Email {
		provider = "email"
		providerID = streamer.Email
	}

	response := dto.UserInfo{
		ID:          streamer.ID,
		DisplayName: streamer.Name,
		Username:    streamer.Name,
		Provider:    provider,
		ProviderID:  providerID,
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
