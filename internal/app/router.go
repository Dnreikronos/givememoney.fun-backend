// internal/app/router.go
package app

import (
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(c *Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RequestLoggerMiddleware(c.Logger))
	router.Use(middleware.ErrorHandlerMiddleware(c.Logger))
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityMiddleware())

	api := router.Group("/api")
	{
		api.POST("/transaction/wallet/:wallet_id", c.Controllers.Transaction.Create)
		api.POST("/transaction", c.Controllers.Transaction.Create)
		api.GET("/ws/alerts/:streamer_id", c.Controllers.Websocket.HandleConnection)
		api.GET("/alerts/:streamer_id", c.Controllers.Alert.ServeAlertPage)

		auth := api.Group("/auth")
		{
			setupTwitchRoutes(auth.Group("/twitch"), c)
			setupKickRoutes(auth.Group("/kick"), c)
			setupAuthRoutes(auth, c)
			setupSessionRoutes(auth, c)
			setupWalletRoutes(auth, c)
			setupTransactionRoutes(auth, c)
		}
	}

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket route for OBS alerts
	router.GET("/ws/streamer/:streamer_id", c.Controllers.Websocket.HandleConnection)

	return router
}

func setupTwitchRoutes(g *gin.RouterGroup, c *Container) {
	g.GET("/login", c.Controllers.Auth.TwitchLogin)
	g.GET("/callback", c.Controllers.Auth.TwitchCallback)
}

func setupKickRoutes(g *gin.RouterGroup, c *Container) {
	g.GET("/login", c.Controllers.Auth.KickLogin)
	g.GET("/callback", c.Controllers.Auth.KickCallback)
}

func setupAuthRoutes(g *gin.RouterGroup, c *Container) {
	g.POST("/register", middleware.AuthRateLimitMiddleware(), c.Controllers.EmailAuth.Register)
	g.POST("/login", middleware.AuthRateLimitMiddleware(), c.Controllers.EmailAuth.Login)
	g.POST("/refresh", c.Controllers.Session.Refresh)
	g.POST("/logout", c.Controllers.Session.Logout)
}

func setupSessionRoutes(g *gin.RouterGroup, c *Container) {
	g.POST("/session", c.Controllers.Session.Create)
	g.GET("/session", c.Controllers.Session.Get)
	g.DELETE("/session", c.Controllers.Session.Delete)
	g.GET("/sessions", c.Controllers.Session.GetActive)
}

func setupWalletRoutes(g *gin.RouterGroup, c *Container) {
	authMiddleware := middleware.AuthMiddleware(c.JWTService)
	g.POST("/wallet", authMiddleware, c.Controllers.Wallet.Create)
	g.GET("/wallet/streamer/:streamer_id", authMiddleware, c.Controllers.Wallet.GetByStreamer)
	g.GET("/wallet/:id", authMiddleware, c.Controllers.Wallet.GetByID)
	g.GET("/wallet/address/:wallet_address", authMiddleware, c.Controllers.Wallet.GetByWalletAddress)
	g.PUT("/wallet/:id", authMiddleware, c.Controllers.Wallet.Update)
	g.DELETE("/wallet/:id", authMiddleware, c.Controllers.Wallet.Delete)
}

func setupTransactionRoutes(g *gin.RouterGroup, c *Container) {
	g.GET("/transaction/wallet/:address_to_id", c.Controllers.Transaction.GetByWalletID)
	g.GET("/transaction/streamer/:streamer_id", c.Controllers.Transaction.GetByStreamerID)
	g.GET("/transaction/:id", c.Controllers.Transaction.GetByID)
	g.GET("/transaction", c.Controllers.Transaction.GetAllTransactions)
}
