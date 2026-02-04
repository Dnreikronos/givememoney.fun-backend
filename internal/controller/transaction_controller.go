package controller

import (
	"context"
	"errors"
	"net/http"

	apperrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionService defines the transaction operations used by TransactionController.
type TransactionService interface {
	Create(ctx context.Context, walletID uuid.UUID, req *model.TransactionRequest) (*model.Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	GetAllTransactions(ctx context.Context) (*[]model.Transaction, error)
	GetByWalletID(ctx context.Context, walletID uuid.UUID) (*[]model.Transaction, error)
	GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error)
}

var _ TransactionService = (*service.TransactionService)(nil)

// TransactionController handles HTTP requests for transaction resources.
type TransactionController struct {
	transactionService TransactionService
}

// NewTransactionController creates a new TransactionController with the given transaction service.
func NewTransactionController(transactionService TransactionService) *TransactionController {
	return &TransactionController{
		transactionService: transactionService,
	}
}

// Create creates a new transaction for a wallet.
func (c *TransactionController) Create(ctx *gin.Context) {
	var req model.TransactionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("Invalid request body", err))
		return
	}

	if req.Currency != "" {
		valid := false
		for _, c := range utils.SupportedCurrencies {
			if req.Currency == c {
				valid = true
				break
			}
		}
		if !valid {
			middleware.AbortWithError(ctx, apperrors.NewValidationError("currency must be one of: ETH, USDT, USDC", nil))
			return
		}
	}

	walletIDStr := ctx.Param("wallet_id")
	if walletIDStr == "" {
		walletIDStr = ctx.Query("wallet_id")
	}
	if walletIDStr == "" {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("wallet_id is required (as URL parameter or query parameter)", nil))
		return
	}

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid wallet_id format", err))
		return
	}

	transaction, err := c.transactionService.Create(ctx, walletID, &req)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to create transaction", err))
		return
	}

	ctx.JSON(http.StatusCreated, transaction)
}

// GetByID returns a transaction by ID.
func (c *TransactionController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid id", err))
		return
	}

	transaction, err := c.transactionService.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.AbortWithError(ctx, apperrors.NewNotFoundError("transaction"))
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get transaction", err))
		return
	}

	ctx.JSON(http.StatusOK, transaction)
}

// GetAllTransactions returns all transactions.
func (c *TransactionController) GetAllTransactions(ctx *gin.Context) {
	transactions, err := c.transactionService.GetAllTransactions(ctx)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get transactions", err))
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

// GetByWalletID returns transactions for a wallet.
func (c *TransactionController) GetByWalletID(ctx *gin.Context) {
	walletID, err := uuid.Parse(ctx.Param("address_to_id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid wallet ID", err))
		return
	}

	transactions, err := c.transactionService.GetByWalletID(ctx, walletID)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get transactions", err))
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

// GetByStreamerID returns transactions for a streamer.
func (c *TransactionController) GetByStreamerID(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("invalid streamer_id", err))
		return
	}

	transactions, err := c.transactionService.GetByStreamerID(ctx, streamerID)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get transactions", err))
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}
