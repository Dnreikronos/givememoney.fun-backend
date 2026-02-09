package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	apperrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QRSettingsServiceInterface interface {
	GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*model.QRSettings, error)
	Upsert(ctx context.Context, streamerID uuid.UUID, req *model.QRSettingsRequest) (*model.QRSettings, error)
}

var _ QRSettingsServiceInterface = (*service.QRSettingsService)(nil)

type QRSettingsController struct {
	service QRSettingsServiceInterface
}

func NewQRSettingsController(service QRSettingsServiceInterface) *QRSettingsController {
	return &QRSettingsController{service: service}
}

func (c *QRSettingsController) Get(ctx *gin.Context) {
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

	settings, err := c.service.GetByStreamerID(ctx, streamerUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, nil)
			return
		}
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to get QR settings", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (c *QRSettingsController) Upsert(ctx *gin.Context) {
	var req model.QRSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(ctx, apperrors.NewValidationError("Invalid request body", err))
		return
	}

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

	settings, err := c.service.Upsert(ctx, streamerUUID, &req)
	if err != nil {
		middleware.AbortWithError(ctx, apperrors.NewInternalError("Failed to save QR settings", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}
