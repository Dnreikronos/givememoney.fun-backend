// internal/app/container.go
package app

import (
	"github.com/Dnreikronos/givememoney.fun-backend/internal/controller"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	Controllers *Controllers
}

type Controllers struct {
	Auth        *controller.AuthController
	EmailAuth   *controller.EmailAuthController
	Session     *controller.SessionController
	Wallet      *controller.WalletController
	Transaction *controller.TransactionController
}

func NewContainer(db *gorm.DB, logger *zap.Logger) (*Container, error) {
	streamerRepo := repository.NewStreamerRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	authService := service.NewAuthService(streamerRepo)
	jwtService := service.NewJWTService()
	sessionService := service.NewSessionService(db, jwtService)
	walletService := service.NewWalletService(walletRepo)
	transactionService := service.NewTransactionService(transactionRepo)

	helpers := controller.NewAuthHelpers(sessionService, streamerRepo)

	controllers := &Controllers{
		Auth:        controller.NewAuthController(authService, jwtService, sessionService, streamerRepo, logger, helpers),
		EmailAuth:   controller.NewEmailAuthController(authService, jwtService, streamerRepo, helpers, logger),
		Session:     controller.NewSessionController(sessionService, jwtService, streamerRepo, helpers, logger),
		Wallet:      controller.NewWalletController(walletService),
		Transaction: controller.NewTransactionController(transactionService),
	}

	return &Container{Controllers: controllers}, nil
}
