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

func (s *WalletService) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	return s.walletRepo.FindByID(ctx, id)
}

func (s *WalletService) GetByHash(ctx context.Context, hash string) (*model.Wallet, error) {
	return s.walletRepo.FindByHash(ctx, hash)
}

func (s *WalletService) GetByStreamerID(ctx context.Context, streamerID uuid.UUID) ([]model.Wallet, error) {
	return s.walletRepo.FindByStreamerID(ctx, streamerID)
}
