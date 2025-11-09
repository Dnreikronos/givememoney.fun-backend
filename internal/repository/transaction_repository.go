package repository

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

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
	if err := tr.db.WithContext(ctx).Where("wallet_id = ?", walletID).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}

func (tr *TransactionRepository) FindByStreamerID(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error) {
	var transactions []model.Transaction
	if err := tr.db.WithContext(ctx).Where("streamer_id = ?", streamerID).Find(&transactions).Error; err != nil {
		return nil, err
	}

	return &transactions, nil
}

// I think that don't make sense to have a update and delete of transactions, because after that is sent by the user the transaction is already on the database and on the chain
// func (tr *TransactionRepository) Update(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error) {
// 	if err := tr.db.WithContext(ctx).Save(transaction).Error; err != nil {
// 		return nil, err
// 	}

// 	return transaction, nil
// }

// func (tr *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
// 	return tr.db.WithContext(ctx).Delete(&model.Transaction{}, "id = ?", id).Error
// }
