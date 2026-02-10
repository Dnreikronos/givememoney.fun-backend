package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/google/uuid"
)

type QRSettingsService struct {
	repo *repository.QRSettingsRepository
}

func NewQRSettingsService(repo *repository.QRSettingsRepository) *QRSettingsService {
	return &QRSettingsService{repo: repo}
}

func (s *QRSettingsService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*model.QRSettings, error) {
	return s.repo.FindByStreamerID(ctx, streamerID)
}

func (s *QRSettingsService) Upsert(ctx context.Context, streamerID uuid.UUID, req *model.QRSettingsRequest) (*model.QRSettings, error) {
	settings := &model.QRSettings{
		StreamerID:      streamerID,
		PixelColor:      req.PixelColor,
		BackgroundColor: req.BackgroundColor,
		BorderColor:     req.BorderColor,
		LogoUrl:         req.LogoUrl,
		LogoSize:        req.LogoSize,
		LogoShape:       req.LogoShape,
		TopText:         req.TopText,
		BottomText:      req.BottomText,
		TextColor:       req.TextColor,
		TextSize:        req.TextSize,
		FrameStyle:      req.FrameStyle,
		QRSize:          req.QRSize,
		PixelPattern:    req.PixelPattern,
	}
	return s.repo.Upsert(ctx, settings)
}
