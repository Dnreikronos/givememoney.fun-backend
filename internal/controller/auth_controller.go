package controller

import (
	"fmt"
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	authService    *service.AuthService
	jwtService     *service.JWTService
	sessionService *service.SessionService
	streamerRepo   interfaces.StreamerRepositoryInterface
	logger         *zap.Logger
	helpers        *AuthHelpers
}

func NewAuthController(
	authService *service.AuthService,
	jwtService *service.JWTService,
	sessionService *service.SessionService,
	streamerRepo interfaces.StreamerRepositoryInterface,
	logger *zap.Logger,
	helpers *AuthHelpers,
) *AuthController {
	return &AuthController{
		authService:    authService,
		jwtService:     jwtService,
		sessionService: sessionService,
		streamerRepo:   streamerRepo,
		logger:         logger,
		helpers:        helpers,
	}
}

func (c *AuthController) TwitchLogin(ctx *gin.Context) {
	authURL, err := c.authService.GetAuthURL(utils.ProviderTwitch)
	if err != nil {
		appErr := errors.NewProviderAPIError("Failed to get Twitch auth URL", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	c.logger.Info("Redirecting to Twitch Auth",
		zap.String("auth_url", authURL),
		zap.String("ip", ctx.ClientIP()),
	)

	ctx.Redirect(http.StatusFound, authURL)
}

func (c *AuthController) KickLogin(ctx *gin.Context) {
	authURL, err := c.authService.GetAuthURL(utils.ProviderKick)
	if err != nil {
		appErr := errors.NewProviderAPIError("Failed  to get Kick auth URL", err)
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

	frontendURL := c.helpers.GetFrontendURL()
	c.logger.Info("Processing Twitch Callback",
		zap.String("frontend_url", frontendURL),
		zap.String("ip", ctx.ClientIP()),
	)

	streamer, err := c.authService.Authenticate(ctx.Request.Context(), utils.ProviderTwitch, req.Code)
	if err != nil {
		c.logger.Error("Twitch authentication failed", zap.Error(err))
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=auth_failed", frontendURL))
		return
	}

	tokenPair, err := c.helpers.CreateSessionForStreamer(ctx, streamer)
	if err != nil {
		c.logger.Error("Session creation failed", zap.Error(err), zap.String("streaner_id", streamer.ID.String()))
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=session_creation_failed", frontendURL))
		return
	}
	c.jwtService.SetTokenCookies(ctx, tokenPair)

	c.logger.Info("Session created sucessfully",
		zap.String("streamer_id", streamer.ID.String()),
		zap.String("provider", string(streamer.Provider)),
	)

	ctx.Redirect(http.StatusFound, fmt.Sprintf("%/dashboard", frontendURL))
}

func (c *AuthController) KickCallback(ctx *gin.Context) {
	var req dto.KickCallbackRequest
	if !middleware.ValidateQuery(ctx, &req) {
		return
	}

	frontendURL := c.helpers.GetFrontendURL()
	c.logger.Info("Processing Kick callback",
		zap.String("frontend_url", frontendURL),
		zap.String("ip", ctx.ClientIP()),
	)
	provider, exists := c.authService.GetProvider(utils.ProviderKick)
	if !exists {
		c.logger.Error("Kick provider not found")
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=provider_not_found", frontendURL))
		return
	}

	kickProvider, ok := provider.(*service.KickProvider)
	if !ok {
		c.logger.Error("Failed to cast provider to KickProvider")
		ctx.Redirect(http.StatusFound, fmt.Sprintf("%s/login?error=invalid_state", frontendURL))
		return
	}

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

	tokenPair, err := c.helpers.CreateSessionForStreamer(ctx, streamer)
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
}
