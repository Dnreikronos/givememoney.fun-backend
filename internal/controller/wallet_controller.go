package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type WalletController struct {
	walletService *service.WalletService
}

func NewWalletController(walletService *service.WalletService) *WalletController {
	return &WalletController{
		walletService: walletService,
	}
}

func (c *WalletController) Create(ctx *gin.Context) {
	var req model.WalletRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(map[string]string)
			for _, e := range validationErrors {
				errors[strings.ToLower(e.Field())] = getValidationError(e)
			}
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":  "Validation failed",
				"fields": errors,
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize wallet address (remove 0x prefix if present, ensure 64 chars)
	walletAddr := strings.TrimSpace(req.WalletAddress)
	if strings.HasPrefix(strings.ToLower(walletAddr), "0x") {
		walletAddr = walletAddr[2:]
	}

	if len(walletAddr) != 64 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":             "wallet_address must be exactly 64 hexadecimal characters",
			"received_length":   len(req.WalletAddress),
			"normalized_length": len(walletAddr),
			"hint":              "Ethereum addresses are 40 hex chars (42 with 0x). You may need a different format or padding.",
		})
		return
	}

	// Validate hex characters
	for _, char := range walletAddr {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "wallet_address must contain only hexadecimal characters (0-9, a-f, A-F)",
			})
			return
		}
	}

	// Store normalized address (lowercase, no 0x)
	req.WalletAddress = strings.ToLower(walletAddr)

	streamerID, exists := ctx.Get("streamer_id")
	if !exists || streamerID == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "streamer_id not found in context - authentication required",
		})
		return
	}

	streamerUUID, ok := streamerID.(uuid.UUID)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid streamer_id type",
			"type":  fmt.Sprintf("%T", streamerID),
		})
		return
	}

	wallet, err := c.walletService.Create(ctx, streamerUUID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "Failed to create wallet",
		})
		return
	}
	ctx.JSON(http.StatusCreated, wallet)
}

func getValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "len":
		return "Must be exactly " + e.Param() + " characters"
	default:
		return "Invalid value"
	}
}

func (c *WalletController) GetByID(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	wallet, err := c.walletService.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "invalid wallet"})
	}

	ctx.JSON(http.StatusOK, wallet)
}

func (c *WalletController) GetByStreamer(ctx *gin.Context) {
	streamerID, err := uuid.Parse(ctx.Param("streamer_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid streamer_id"})
		return
	}

	wallets, err := c.walletService.GetByStreamerID(ctx, streamerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, wallets)
}

func (c *WalletController) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
	}

	var req model.WalletUpdateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusOK, id)
}

func (c *WalletController) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.walletService.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}
