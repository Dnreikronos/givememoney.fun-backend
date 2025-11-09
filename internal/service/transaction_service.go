package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/google/uuid"
)

type TransactionService struct {
	transactionRepo *repository.TransactionRepository
}

func NewTransactionService(transactionRepo *repository.TransactionRepository) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
	}
}
