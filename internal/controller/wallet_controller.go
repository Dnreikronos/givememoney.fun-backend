package controller

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	streamerID, _ := ctx.Get("streamer_id")
	wallet, err := c.walletService.Create(ctx, streamerID.(uuid.UUID), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, wallet)
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
