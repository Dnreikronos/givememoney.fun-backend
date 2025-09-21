package main

import (
	"log"
	"os"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/config"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/controller"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/database/connection"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
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

	err = db.AutoMigrate(&model.Wallet{}, &model.Streamer{}, &model.Session{}, &model.RefreshToken{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed successfully")

	// Initialize logger
	loggerService, err := service.NewLoggerService()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	logger := loggerService.GetLogger()
	defer loggerService.Sync()

	// Initialize repositories
	streamerRepo := repository.NewStreamerRepository(db)

	// Initialize services
	authService := service.NewAuthService(streamerRepo)
	jwtService := service.NewJWTService()
	sessionService := service.NewSessionService(db, jwtService)

	// Initialize controllers
	authController := controller.NewAuthController(authService, jwtService, sessionService, streamerRepo, logger)

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
				twitch.POST("/token", authController.TwitchToken)
				twitch.GET("/user", authController.TwitchUser)
			}

			// General auth endpoints
			auth.POST("/refresh", authController.RefreshToken)
			auth.POST("/logout", authController.Logout)

			// Session management endpoints
			auth.POST("/session", authController.CreateSession)
			auth.GET("/session", authController.GetSession)
			auth.DELETE("/session", authController.DeleteSession)
			auth.GET("/sessions", authController.GetActiveSessions)
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
