package connection

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig holds database configuration
type DBConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// getDBConfig returns database configuration from environment variables with defaults
func getDBConfig() DBConfig {
	config := DBConfig{
		MaxIdleConns:    10,  // Default idle connections
		MaxOpenConns:    25,  // Default max connections
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}

	if val, exists := os.LookupEnv("DB_MAX_IDLE_CONNS"); exists {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxIdleConns = parsed
		}
	}

	if val, exists := os.LookupEnv("DB_MAX_OPEN_CONNS"); exists {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.MaxOpenConns = parsed
		}
	}

	if val, exists := os.LookupEnv("DB_CONN_MAX_LIFETIME_MINUTES"); exists {
		if parsed, err := strconv.Atoi(val); err == nil {
			config.ConnMaxLifetime = time.Duration(parsed) * time.Minute
		}
	}

	return config
}

func OpenConnection() (*gorm.DB, error) {
	dsn := os.Getenv("POSTGRES_SOURCE")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("POSTGRES_TIME_ZONE"),
		)
	}

	// Configure GORM with optimized settings
	config := &gorm.Config{
		PrepareStmt: true, // Enable prepared statements for better performance
		Logger:      logger.Default.LogMode(logger.Error), // Only log errors in production
	}

	db, err := gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Apply connection pool configuration
	dbConfig := getDBConfig()
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbConfig.ConnMaxIdleTime)

	// Test the connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to the database with pool config: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%v",
		dbConfig.MaxIdleConns, dbConfig.MaxOpenConns, dbConfig.ConnMaxLifetime)

	return db, nil
}
