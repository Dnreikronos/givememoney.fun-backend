package controller

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EmailAuthController handles email-based registration and login.
type EmailAuthController struct {
	authService  *service.AuthService
	jwtService   *service.JWTService
	streamerRepo interfaces.StreamerRepositoryInterface
	logger       *zap.Logger
	helpers      *AuthHelpers
}

func NewEmailAuthController(
	authService *service.AuthService,
	jwtService *service.JWTService,
	streamerRepo interfaces.StreamerRepositoryInterface,
	helpers *AuthHelpers,
	logger *zap.Logger,
) *EmailAuthController {
	return &EmailAuthController{
		authService:  authService,
		jwtService:   jwtService,
		streamerRepo: streamerRepo,
		helpers:      helpers,
		logger:       logger,
	}
}

// Register registers a new streamer with email and password.
func (c *EmailAuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	if !middleware.ValidateRegisterRequest(ctx, &req) {
		return
	}

	c.logger.Info("Processing email registration",
		zap.String("email", req.Email),
		zap.String("name", req.Name),
		zap.String("ip", ctx.ClientIP()),
	)

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

	streamer, err := c.authService.RegisterWithEmail(ctx.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		c.logger.Error("Email registration failed", zap.Error(err))
		appErr := errors.NewInternalError("Registration failed", err)
		middleware.AbortWithError(ctx, appErr)
		return
	}

	tokenPair, err := c.helpers.CreateSessionForStreamer(ctx, streamer)
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

	userResponse := c.helpers.GetUserResponse(streamer)
	ctx.JSON(http.StatusCreated, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         userResponse,
	})
}

// Login authenticates a streamer with email and password.
func (c *EmailAuthController) Login(ctx *gin.Context) {
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

	tokenPair, err := c.helpers.CreateSessionForStreamer(ctx, streamer)
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

	userResponse := c.helpers.GetUserResponse(streamer)
	ctx.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
		User:         userResponse,
	})
}
