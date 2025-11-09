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

func (s *TransactionService) Create(ctx context.Context, walletID uuid.UUID, req *model.TransactionRequest) (*model.Transaction, error) {
	transaction :=
		&model.Transaction{
			Amount:      req.Amount,
			Message:     req.Message,
			TxHash:      req.TxHash,
			AddressFrom: req.AddressFrom,
			AddressToID: walletID,
		}
	return s.transactionRepo.Create(ctx, transaction)
}

func (s *TransactionService) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	return s.transactionRepo.FindByID(ctx, id)
}

func (s *TransactionService) GetAllTransactions(ctx context.Context) (*[]model.Transaction, error) {
	return s.transactionRepo.FindAll(ctx)
}

func (s *TransactionService) GetByWalletID(ctx context.Context, walletID uuid.UUID) (*[]model.Transaction, error) {
	return s.transactionRepo.FindByWalletID(ctx, walletID)
}

func (s *TransactionService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) (*[]model.Transaction, error) {
	return s.transactionRepo.FindByStreamerID(ctx, streamerID)
}
