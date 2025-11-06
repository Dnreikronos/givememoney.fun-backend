package main

import (
	"log"
	"os"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/app"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/config"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/database/connection"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
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
	defer loggerService.Sync()

	container, err := app.NewContainer(db, loggerService.GetLogger())
	if err != nil {
		log.Fatal("Failed to initialize container:", err)
	}

	router := app.SetupRouter(container)

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
