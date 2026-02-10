package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertSettingsRepository struct {
	db *gorm.DB
}

func NewAlertSettingsRepository(db *gorm.DB) *AlertSettingsRepository {
	return &AlertSettingsRepository{db: db}
}

func (r *AlertSettingsRepository) FindByStreamerID(ctx context.Context, streamerID uuid.UUID) (*model.AlertSettings, error) {
	var settings model.AlertSettings
	if err := r.db.WithContext(ctx).Where("streamer_id = ?", streamerID).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *AlertSettingsRepository) Upsert(ctx context.Context, settings *model.AlertSettings) (*model.AlertSettings, error) {
	var existing model.AlertSettings
	err := r.db.WithContext(ctx).Where("streamer_id = ?", settings.StreamerID).First(&existing).Error

	if err == nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
		if err := r.db.WithContext(ctx).Save(settings).Error; err != nil {
			return nil, err
		}
		return settings, nil
	}

	if err := r.db.WithContext(ctx).Create(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}
