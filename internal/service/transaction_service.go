package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/websocket"
	"github.com/google/uuid"
)

// TransactionService handles transaction business logic and real-time broadcast.
type TransactionService struct {
	transactionRepo *repository.TransactionRepository
	walletRepo      *repository.WalletRepository
	hub             *websocket.Hub
}

// NewTransactionService creates a new TransactionService with the given dependencies.
func NewTransactionService(transactionRepo *repository.TransactionRepository, walletRepo *repository.WalletRepository, hub *websocket.Hub) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		walletRepo:      walletRepo,
		hub:             hub,
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
	result, err := s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, err
	}

	wallet, err := s.walletRepo.FindByID(ctx, walletID)
	if err == nil && wallet != nil {
		s.hub.BroadcastToStreamer(wallet.StreamerID, result)
	}

	return result, nil
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
