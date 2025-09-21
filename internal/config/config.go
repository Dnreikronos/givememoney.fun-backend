package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
	return nil
}

func GetServerPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}
	return port
}
