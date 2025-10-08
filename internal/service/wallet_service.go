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
		Hash:       req.Hash,
		StreamerID: streamerID,
	}
	return s.walletRepo.Create(ctx, wallet)
}
