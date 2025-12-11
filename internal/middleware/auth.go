package middleware

import (
	"strings"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			appErr := errors.NewUnauthorizedError("Authorization header required")
			AbortWithError(c, appErr)
			return
		}

		// Extract Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			appErr := errors.NewUnauthorizedError("Invalid authorization header format. Expected: Bearer <token>")
			AbortWithError(c, appErr)
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			appErr := errors.NewUnauthorizedError("Token is empty")
			AbortWithError(c, appErr)
			return
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			appErr := errors.NewUnauthorizedError("Invalid or expired token: " + err.Error())
			AbortWithError(c, appErr)
			return
		}

		// Set streamer_id in context
		c.Set("streamer_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_provider", claims.Provider)

		c.Next()
	}
}
