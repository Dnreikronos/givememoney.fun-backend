package controller

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	transactionService *service.TransactionService
}

func NewTransactionController(transactionService *service.TransactionService) *TransactionController {
	return &TransactionController{
		transactionService: transactionService,
	}
}

func (c *TransactionController) Create(ctx *gin.Context) {
	var req model.TransactionRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get wallet ID from URL parameter or request body
	walletIDStr := ctx.Param("wallet_id")
	if walletIDStr == "" {
		// Try to get from query parameter
		walletIDStr = ctx.Query("wallet_id")
	}
	
	if walletIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "wallet_id is required (as URL parameter or query parameter)"})
		return
	}

	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet_id format"})
		return
	}

	transaction, err := c.transactionService.Create(ctx, walletID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, transaction)
}

func (c *TransactionController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	transaction, err := c.transactionService.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transaction)
}

func (c *TransactionController) GetAllTransactions(ctx *gin.Context) {
	transactions, err := c.transactionService.GetAllTransactions(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

func (c *TransactionController) GetByWalletID(ctx *gin.Context) {
	walletID, err := uuid.Parse(ctx.Param("address_to_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	transactions, err := c.transactionService.GetByWalletID(ctx, walletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}

func (c *TransactionController) GetByStreamerID(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	transactions, err := c.transactionService.GetByStreamerID(ctx, streamerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, transactions)
}
