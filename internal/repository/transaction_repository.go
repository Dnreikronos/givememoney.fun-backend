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
		re
		turn nil, err
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
