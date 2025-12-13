package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WalletRepository struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *model.Wallet) (*model.Wallet, error) {
	if err := r.db.WithContext(ctx).Create(wallet).Error; err != nil {
		return nil, err
	}

	// Reload wallet with Streamer relationship preloaded
	var loadedWallet model.Wallet
	if err := r.db.WithContext(ctx).Preload("Streamer").First(&loadedWallet, wallet.ID).Error; err != nil {
		return nil, err
	}

	return &loadedWallet, nil
}

func (r *WalletRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := r.db.WithContext(ctx).First(&wallet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) FindByWalletAddress(ctx context.Context, walletAddress string) (*model.Wallet, error) {
	var wallet model.Wallet
	if err := r.db.WithContext(ctx).First(&wallet, "wallet_address = ?", walletAddress).Error; err != nil {
		return nil, err
	}
	return &wallet, nil
}

func (r *WalletRepository) FindByStreamerID(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error) {
	var wallets []model.Wallet
	if err := r.db.WithContext(ctx).Where("streamer_id = ?", streamerID).Find(&wallets).Error; err != nil {
		return nil, err
	}
	return wallets, nil
}

func (r *WalletRepository) Update(ctx context.Context, wallet *model.Wallet) (*model.Wallet, error) {
	if err := r.db.WithContext(ctx).Save(wallet).Error; err != nil {
		return nil, err
	}
	return wallet, nil
}

func (r *WalletRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Wallet{}, "id = ?", id).Error
}
