package connection

import (
	"log"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/model"
	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) {
	CreateTable(db)
}
func CreateTable(db *gorm.DB) {
	err := db.AutoMigrate(
		&model.Wallet{},
		&model.Streamer{},
		&model.Session{},
		&model.RefreshToken{},
		&model.Transaction{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
