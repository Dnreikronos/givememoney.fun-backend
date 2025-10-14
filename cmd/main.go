package main

import (
	"log"
	"os"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/config"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/controller"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/database/connection"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/repository"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	if err := config.Load(); err != nil {
		log.Println("Warning: Failed to load config file, using environment variables")
	}

	db, err := connection.OpenConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	connection.RunMigration(db)

	loggerService, err := service.NewLoggerService()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	logger := loggerService.GetLogger()
	defer loggerService.Sync()

	streamerRepo := repository.NewStreamerRepository(db)

	authService := service.NewAuthService(streamerRepo)
	jwtService := service.NewJWTService()
	sessionService := service.NewSessionService(db, jwtService)

	helpers := controller.NewAuthHelpers(sessionService, streamerRepo)
	authController := controller.NewAuthController(authService, jwtService, sessionService, streamerRepo, logger, helpers)
	emailAuthController := controller.NewEmailAuthController(authService, jwtService, streamerRepo, helpers, logger)
	sessionController := controller.NewSessionController(sessionService, jwtService, streamerRepo, helpers, logger)

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityMiddleware())

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			twitch := auth.Group("/twitch")
			{
				twitch.GET("/login", authController.TwitchLogin)
				twitch.GET("/callback", authController.TwitchCallback)
				twitch.GET("/user", authController.TwitchUser)
			}

			kick := auth.Group("/kick")
			{
				kick.GET("/login", authController.KickLogin)
				kick.GET("/callback", authController.KickCallback)
				kick.GET("/user", authController.KickUser)
			}

			auth.POST("/register", middleware.AuthRateLimitMiddleware(), emailAuthController.Register)
			auth.POST("/login", middleware.AuthRateLimitMiddleware(), emailAuthController.Login)
			auth.POST("/refresh", sessionController.Refresh)
			auth.POST("/logout", sessionController.Logout)
			auth.POST("/session", sessionController.Create)
			auth.GET("/session", sessionController.Get)
			auth.DELETE("/session", sessionController.Delete)
			auth.GET("/sessions", sessionController.GetActive)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := config.GetServerPort()
	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = "9090"
		}
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
