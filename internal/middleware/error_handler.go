package middleware

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			if appErr, ok := errors.IsAppError(err); ok {
				handleAppError(c, appErr, logger)
			} else {
				handleGenericError(c, err, logger)
			}
		}
	}
}

func handleAppError(c *gin.Context, appErr *errors.AppError, logger *zap.Logger) {
	fields := []zap.Field{
		zap.String("error_code", string(appErr.Code)),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
	}

	if requestID := c.GetString("request_id"); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if len(appErr.Context) > 0 {
		fields = append(fields, zap.Any("context", appErr.Context))
	}

	switch appErr.StatusCode {
	case http.StatusInternalServerError:
		if appErr.Err != nil {
			fields = append(fields, zap.Error(appErr.Err))
		}
		logger.Error("Application error", fields...)
	case http.StatusUnauthorized, http.StatusForbidden:
		logger.Warn("Authentication/Authorization error", fields...)
	case http.StatusBadRequest, http.StatusNotFound, http.StatusConflict:
		logger.Info("Client error", fields...)
	default:
		logger.Info("Application error", fields...)
	}

	c.JSON(appErr.StatusCode, appErr.ToErrorResponse())
}

func handleGenericError(c *gin.Context, err error, logger *zap.Logger) {
	logger.Error("Unhandled error",
		zap.Error(err),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()),
		zap.String("request_id", c.GetString("request_id")),
	)

	c.JSON(http.StatusInternalServerError, errors.ErrorResponse{
		Error:   "internal_error",
		Message: "An internal error occurred",
		Code:    errors.ErrorCodeInternalError,
	})
}

func AbortWithError(c *gin.Context, appErr *errors.AppError) {
	c.Error(appErr)
	c.Abort()
}

func AbortWithValidationError(c *gin.Context, message string, err error) {
	appErr := errors.NewValidationError(message, err)
	AbortWithError(c, appErr)
}

func AbortWithNotFoundError(c *gin.Context, resource string) {
	appErr := errors.NewNotFoundError(resource)
	AbortWithError(c, appErr)
}

func AbortWithUnauthorizedError(c *gin.Context, message string) {
	appErr := errors.NewUnauthorizedError(message)
	AbortWithError(c, appErr)
}

func AbortWithDatabaseError(c *gin.Context, operation string, err error) {
	appErr := errors.NewDatabaseError(operation, err)
	AbortWithError(c, appErr)
}

func AbortWithInternalError(c *gin.Context, message string, err error) {
	appErr := errors.NewInternalError(message, err)
	AbortWithError(c, appErr)
}
