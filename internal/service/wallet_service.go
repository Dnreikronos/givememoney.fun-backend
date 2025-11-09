package service

import (
	"context"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"

	"github.com/google/uuid"
)

type WalletService struct {
	walletRepo *repository.WalletRepository
}

func NewWalletService(walletRepo *repository.WalletRepository) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

func (s *WalletService) Create(ctx context.Context, streamerID uuid.UUID, req *model.WalletRequest) (*model.Wallet, error) {
	wallet := &model.Wallet{
		WalletAddress:  req.WalletAddress,
		StreamerID:     streamerID,
		WalletProvider: req.WalletProvider,
	}
	return s.walletRepo.Create(ctx, wallet)
}

func (s *WalletService) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	return s.walletRepo.FindByID(ctx, id)
}

func (s *WalletService) GetByWalletAddress(ctx context.Context, WalletAddress string) (*model.Wallet, error) {
	return s.walletRepo.FindByWalletAddress(ctx, WalletAddress)
}

func (s *WalletService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error) {
	return s.walletRepo.FindByStreamerID(ctx, streamerID)
}

func (s *WalletService) Update(ctx context.Context, id uuid.UUID, req *model.WalletUpdateInput) (*model.Wallet, error) {
	wallet, err := s.walletRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	wallet.WalletAddress = req.WalletAddress
	return s.walletRepo.Update(ctx, wallet)
}

func (s *WalletService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.walletRepo.Delete(ctx, id)
}
