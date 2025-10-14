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

func newProductionLogger() (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

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

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"

	return config.Build()
}

func newDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config.Build()
}

func (ls *LoggerService) GetLogger() *zap.Logger {
	return ls.logger
}

func (ls *LoggerService) Info(msg string, fields ...zap.Field) {
	ls.logger.Info(msg, fields...)
}

func (ls *LoggerService) Error(msg string, fields ...zap.Field) {
	ls.logger.Error(msg, fields...)
}

func (ls *LoggerService) Warn(msg string, fields ...zap.Field) {
	ls.logger.Warn(msg, fields...)
}

func (ls *LoggerService) Debug(msg string, fields ...zap.Field) {
	ls.logger.Debug(msg, fields...)
}

func (ls *LoggerService) Fatal(msg string, fields ...zap.Field) {
	ls.logger.Fatal(msg, fields...)
}

func (ls *LoggerService) With(fields ...zap.Field) *LoggerService {
	return &LoggerService{
		logger: ls.logger.With(fields...),
	}
}

func (ls *LoggerService) Sync() error {
	return ls.logger.Sync()
}

func (ls *LoggerService) Close() error {
	return ls.Sync()
}
