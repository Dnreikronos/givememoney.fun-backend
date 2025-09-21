package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/interfaces"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StreamerRepository struct {
	db *gorm.DB
}

var _ interfaces.StreamerRepositoryInterface = (*StreamerRepository)(nil)

func NewStreamerRepository(db *gorm.DB) interfaces.StreamerRepositoryInterface {
	return &StreamerRepository{db: db}
}

func (r *StreamerRepository) FindByProvider(ctx context.Context, provider utils.StreamerProvider, providerID string) (*model.Streamer, error) {
	var streamer model.Streamer
	err := r.db.WithContext(ctx).Preload("Wallet").Where("provider = ? AND provider_id = ?", provider, providerID).First(&streamer).Error
	if err != nil {
		return nil, err
	}
	return &streamer, nil
}

func (r *StreamerRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Streamer, error) {
	var streamer model.Streamer
	err := r.db.WithContext(ctx).Preload("Wallet").First(&streamer, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &streamer, nil
}

func (r *StreamerRepository) FindByEmail(ctx context.Context, email string) (*model.Streamer, error) {
	var streamer model.Streamer
	err := r.db.WithContext(ctx).Preload("Wallet").Where("email = ?", email).First(&streamer).Error
	if err != nil {
		return nil, err
	}
	return &streamer, nil
}

func (r *StreamerRepository) Create(ctx context.Context, streamer *model.Streamer) error {
	return r.db.WithContext(ctx).Create(streamer).Error
}

func (r *StreamerRepository) Update(ctx context.Context, streamer *model.Streamer) error {
	return r.db.WithContext(ctx).Save(streamer).Error
}

func (r *StreamerRepository) CreateWithWallet(ctx context.Context, streamer *model.Streamer, wallet *model.Wallet) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(wallet).Error; err != nil {
			return err
		}

		streamer.WalletID = wallet.ID
		if err := tx.Create(streamer).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *StreamerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Streamer{}, id).Error
}
