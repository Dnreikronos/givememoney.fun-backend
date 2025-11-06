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
	Auth      *controller.AuthController
	EmailAuth *controller.EmailAuthController
	Session   *controller.SessionController
	Wallet    *controller.WalletController
}

func NewContainer(db *gorm.DB, logger *zap.Logger) (*Container, error) {
	streamerRepo := repository.NewStreamerRepository(db)
	walletRepo := repository.NewWalletRepository(db)

	authService := service.NewAuthService(streamerRepo)
	jwtService := service.NewJWTService()
	sessionService := service.NewSessionService(db, jwtService)
	walletService := service.NewWalletService(walletRepo)

	helpers := controller.NewAuthHelpers(sessionService, streamerRepo)

	controllers := &Controllers{
		Auth:      controller.NewAuthController(authService, jwtService, sessionService, streamerRepo, logger, helpers),
		EmailAuth: controller.NewEmailAuthController(authService, jwtService, streamerRepo, helpers, logger),
		Session:   controller.NewSessionController(sessionService, jwtService, streamerRepo, helpers, logger),
		Wallet:    controller.NewWalletController(walletService),
	}

	return &Container{Controllers: controllers}, nil
}
