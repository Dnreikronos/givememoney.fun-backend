// internal/app/container.go
package app

import (
	"github.com/Dnreikronos/givememoney.fun-backend/internal/controller"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	Controllers *Controllers
	Hub         *websocket.Hub
	JWTService  *service.JWTService
}

type Controllers struct {
	Auth          *controller.AuthController
	EmailAuth     *controller.EmailAuthController
	Session       *controller.SessionController
	Wallet        *controller.WalletController
	Transaction   *controller.TransactionController
	Websocket     *controller.WebsocketController
	Alert         *controller.AlertController
	AlertSettings *controller.AlertSettingsController
	QRSettings    *controller.QRSettingsController
}

func NewContainer(db *gorm.DB, logger *zap.Logger) (*Container, error) {

	hub := websocket.NewHub()
	go hub.Run()

	streamerRepo := repository.NewStreamerRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	alertSettingsRepo := repository.NewAlertSettingsRepository(db)
	qrSettingsRepo := repository.NewQRSettingsRepository(db)

	authService := service.NewAuthService(streamerRepo)
	jwtService := service.NewJWTService()
	sessionService := service.NewSessionService(db, jwtService)
	walletService := service.NewWalletService(walletRepo)
	transactionService := service.NewTransactionService(transactionRepo, walletRepo, hub)
	alertSettingsService := service.NewAlertSettingsService(alertSettingsRepo)
	qrSettingsService := service.NewQRSettingsService(qrSettingsRepo)

	helpers := controller.NewAuthHelpers(sessionService, streamerRepo)

	controllers := &Controllers{
		Auth:        controller.NewAuthController(authService, jwtService, sessionService, streamerRepo, logger, helpers),
		EmailAuth:   controller.NewEmailAuthController(authService, jwtService, streamerRepo, helpers, logger),
		Session:     controller.NewSessionController(sessionService, jwtService, streamerRepo, helpers, logger),
		Wallet:        controller.NewWalletController(walletService),
		Transaction:   controller.NewTransactionController(transactionService),
		Websocket:     controller.NewWebsocketController(hub),
		Alert:         controller.NewAlertController(alertSettingsService, hub),
		AlertSettings: controller.NewAlertSettingsController(alertSettingsService),
		QRSettings:    controller.NewQRSettingsController(qrSettingsService),
	}

	return &Container{
		Controllers: controllers,
		Hub:         hub,
		JWTService:  jwtService,
	}, nil
}
