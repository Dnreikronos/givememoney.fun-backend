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

	return wallet, nil
}
