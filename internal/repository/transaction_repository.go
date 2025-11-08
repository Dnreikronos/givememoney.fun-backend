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

