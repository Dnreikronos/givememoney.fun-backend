// internal/app/router.go
package app

import (
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter(c *Container) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityMiddleware())

	api := router.Group("/api")
	{
		api.POST("/transaction", c.Controllers.Transaction.Create)

		auth := api.Group("/auth")
		{
			setupTwitchRoutes(auth.Group("/twitch"), c)
			setupKickRoutes(auth.Group("/kick"), c)
			setupAuthRoutes(auth, c)
			setupSessionRoutes(auth, c)
			setupWalletRoutes(auth, c)
		}
	}

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

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
	g.POST("/wallet", c.Controllers.Wallet.Create)
	g.GET("/wallet/streamer/:streamer_id", c.Controllers.Wallet.GetByStreamer)
	g.GET("/wallet/:id", c.Controllers.Wallet.GetByID)
	g.PUT("/wallet/:id", c.Controllers.Wallet.Update)
	g.DELETE("/wallet/:id", c.Controllers.Wallet.Delete)
}
