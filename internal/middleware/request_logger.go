package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		}

		if streamerID := c.Param("streamer_id"); streamerID != "" {
			fields = append(fields, zap.String("streamer_id", streamerID))
		}

		switch {
		case status >= 500:
			logger.Error("HTTP request", fields...)
		case status >= 400:
			logger.Warn("HTTP request", fields...)
		default:
			logger.Info("HTTP request", fields...)
		}
	}
}
