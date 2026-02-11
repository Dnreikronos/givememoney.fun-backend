package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apperrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/validator"
	"github.com/gin-gonic/gin"
	pv "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// WalletService defines the wallet operations used by WalletController.
// Accepting an interface enables testing with mocks (see interfaces.md).
type WalletService interface {
	Create(ctx context.Context, streamerID uuid.UUID, req *model.WalletRequest) (*model.Wallet, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error)
	GetByWalletAddress(ctx context.Context, walletAddress string) (*model.Wallet, error)
	GetByStreamerID(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error)
	Update(ctx context.Context, id uuid.UUID, req *model.WalletUpdateInput) (*model.Wallet, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Ensure *service.WalletService satisfies WalletService at compile time.
var _ WalletService = (*service.WalletService)(nil)

// WalletController handles HTTP requests for wallet resources.
type WalletController struct {
	walletService WalletService
}

// NewWalletController creates a new WalletController with the given wallet service.
func NewWalletController(walletService WalletService) *WalletController {
	return &WalletController{
		walletService: walletService,
	}
}

// Create creates a new wallet for the authenticated streamer.
func (c *WalletController) Create(ctx *gin.Context) {
	var req model.WalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(pv.ValidationErrors); ok {
			fieldErrs := make(map[string]string)
			for _, e := range validationErrors {
				fieldErrs[strings.ToLower(e.Field())] = getValidationError(e)
			}
			appErr := apperrors.NewValidationError("Validation failed", err).WithContext("fields", fieldErrs)
			middleware.AbortWithError(ctx, appErr)
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewValidationError("Invalid request body", err))
		return
	}

	normalizedAddr, valErr := validator.NormalizeAndValidateWalletAddress(req.WalletProvider, req.WalletAddress)
	if valErr != nil {
		middleware.AbortWithError(ctx, valErr)
		return
	}
	req.WalletAddress = normalizedAddr

	streamerID, exists := ctx.Get("streamer_id")
	if !exists || streamerID == nil {
		middleware.AbortWithError(ctx, apperrors.NewUnauthorizedError("streamer_id not found in context - authentication required"))
		return
	}

	streamerUUID, ok := streamerID.(uuid.UUID)
	if !ok {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("invalid streamer_id type", nil).
			WithContext("type", fmt.Sprintf("%T", streamerID)))
		return
	}

	wallet, err := c.walletService.Create(ctx, streamerUUID, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			middleware.AbortWithError(ctx, apperrors.NewAppError(apperrors.ErrorCodeConflict, "Wallet address already registered", err))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to create wallet", err))
		return
	}
	ctx.JSON(http.StatusCreated, wallet)
}

// getValidationError maps go-playground validator tags to user-facing messages.
func getValidationError(e pv.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "len":
		return "Must be exactly " + e.Param() + " characters"
	default:
		return "Invalid value"
	}
}

// GetByID returns a wallet by ID.
func (c *WalletController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid id", err))
		return
	}

	wallet, err := c.walletService.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithError(ctx, apperrors.NewNotFoundError("wallet"))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get wallet", err))
		return
	}

	ctx.JSON(http.StatusOK, wallet)
}

// GetByWalletAddress returns a wallet by its address.
func (c *WalletController) GetByWalletAddress(ctx *gin.Context) {
	walletAddress := ctx.Param("wallet_address")
	wallet, err := c.walletService.GetByWalletAddress(ctx, walletAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithError(ctx, apperrors.NewNotFoundError("wallet"))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get wallet", err))
		return
	}
	ctx.JSON(http.StatusOK, wallet)
}

// GetByStreamer returns all wallets for a streamer.
func (c *WalletController) GetByStreamer(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid streamer_id", err))
		return
	}

	wallets, err := c.walletService.GetByStreamerID(ctx, streamerID)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get wallets", err))
		return
	}

	ctx.JSON(http.StatusOK, wallets)
}

// Update updates an existing wallet by ID.
func (c *WalletController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid id", err))
		return
	}

	var req model.WalletUpdateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("Invalid request body", err))
		return
	}

	wallet, err := c.walletService.Update(ctx, id, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithError(ctx, apperrors.NewNotFoundError("wallet"))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to update wallet", err))
		return
	}

	ctx.JSON(http.StatusOK, wallet)
}

// GetByIDPublic returns public wallet information (no auth required).
func (c *WalletController) GetByIDPublic(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid id", err))
		return
	}

	wallet, err := c.walletService.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithError(ctx, apperrors.NewNotFoundError("wallet"))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get wallet", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":              wallet.ID,
		"wallet_provider": wallet.WalletProvider,
		"wallet_address":  wallet.WalletAddress,
		"streamer_id":     wallet.StreamerID,
		"streamer_name":   wallet.Streamer.Name,
	})
}

// Delete deletes a wallet by ID.
func (c *WalletController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid id", err))
		return
	}

	if err := c.walletService.Delete(ctx, id); err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to delete wallet", err))
		return
	}

	ctx.Status(http.StatusOK)
}
