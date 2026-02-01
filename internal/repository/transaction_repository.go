package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionRepository handles transaction persistence.
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new TransactionRepository with the given DB.
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (tr *TransactionRepository) Create(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
	if err := tr.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return nil, err
	}
	return transaction, nil
}

func (tr *TransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	var transaction model.Transaction
	if err := tr.db.WithContext(ctx).First(&transaction, id).Error; err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (tr *TransactionRepository) FindAll(ctx context.Context) (*[]model.Transaction, error) {
	var transactions []model.Transaction
	if err := tr.db.WithContext(ctx).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}

func (tr *TransactionRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID) (*[]model.Transaction, error) {
	var transactions []model.Transaction
	if err := tr.db.WithContext(ctx).Where("address_to_id = ?", walletID).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}

func (tr *TransactionRepository) FindByStreamerID(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error) {
	var transactions []model.Transaction
	// Join with wallets table to filter by streamer_id
	if err := tr.db.WithContext(ctx).
		Joins("JOIN wallets ON transactions.address_to_id = wallets.id").
		Where("wallets.streamer_id = ?", streamerID).
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}
