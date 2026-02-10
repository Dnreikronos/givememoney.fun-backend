package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QRSettingsRepository struct {
	db *gorm.DB
}

func NewQRSettingsRepository(db *gorm.DB) *QRSettingsRepository {
	return &QRSettingsRepository{db: db}
}

func (r *QRSettingsRepository) FindByStreamerID(ctx context.Context, streamerID uuid.UUID) (*model.QRSettings, error) {
	var settings model.QRSettings
	if err := r.db.WithContext(ctx).Where("streamer_id = ?", streamerID).First(&settings).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *QRSettingsRepository) Upsert(ctx context.Context, settings *model.QRSettings) (*model.QRSettings, error) {
	var existing model.QRSettings
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
