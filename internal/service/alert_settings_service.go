package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/google/uuid"
)

type AlertSettingsService struct {
	repo *repository.AlertSettingsRepository
}

func NewAlertSettingsService(repo *repository.AlertSettingsRepository) *AlertSettingsService {
	return &AlertSettingsService{repo: repo}
}

func (s *AlertSettingsService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*model.AlertSettings, error) {
	return s.repo.FindByStreamerID(ctx, streamerID)
}

func (s *AlertSettingsService) Upsert(ctx context.Context, streamerID uuid.UUID, req *model.AlertSettingsRequest) (*model.AlertSettings, error) {
	settings := &model.AlertSettings{
		StreamerID:      streamerID,
		BackgroundColor: req.BackgroundColor,
		TextColor:       req.TextColor,
		MessageColor:    req.MessageColor,
		AccentColor:     req.AccentColor,
		HeaderText:      req.HeaderText,
		ShowDonorName:   req.ShowDonorName,
		ShowAmount:      req.ShowAmount,
		ShowMessage:     req.ShowMessage,
		MinDuration:     req.MinDuration,
		MaxDuration:     req.MaxDuration,
		SoundEnabled:    req.SoundEnabled,
		Position:        req.Position,
	}
	return s.repo.Upsert(ctx, settings)
}
