package service

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerService manages structured logging
type LoggerService struct {
	logger *zap.Logger
}

// NewLoggerService creates a new logger service with production or development configuration
func NewLoggerService() (*LoggerService, error) {
	var logger *zap.Logger
	var err error

	env := os.Getenv("GO_ENV")
	if env == "production" {
		logger, err = newProductionLogger()
	} else {
		logger, err = newDevelopmentLogger()
	}

	if err != nil {
		return nil, err
	}

	return &LoggerService{logger: logger}, nil
}

// newProductionLogger creates a production-ready logger
func newProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Configure output
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Configure encoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"

	return config.Build()
}

// newDevelopmentLogger creates a development-friendly logger
func newDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config.Build()
}

// GetLogger returns the underlying zap logger
func (ls *LoggerService) GetLogger() *zap.Logger {
	return ls.logger
}

// Info logs an info message
func (ls *LoggerService) Info(msg string, fields ...zap.Field) {
	ls.logger.Info(msg, fields...)
}

// Error logs an error message
func (ls *LoggerService) Error(msg string, fields ...zap.Field) {
	ls.logger.Error(msg, fields...)
}

// Warn logs a warning message
func (ls *LoggerService) Warn(msg string, fields ...zap.Field) {
	ls.logger.Warn(msg, fields...)
}

// Debug logs a debug message
func (ls *LoggerService) Debug(msg string, fields ...zap.Field) {
	ls.logger.Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func (ls *LoggerService) Fatal(msg string, fields ...zap.Field) {
	ls.logger.Fatal(msg, fields...)
}

// With creates a child logger with additional fields
func (ls *LoggerService) With(fields ...zap.Field) *LoggerService {
	return &LoggerService{
		logger: ls.logger.With(fields...),
	}
}

// Sync flushes any buffered log entries
func (ls *LoggerService) Sync() error {
	return ls.logger.Sync()
}

// Close properly closes the logger
func (ls *LoggerService) Close() error {
	return ls.Sync()
}