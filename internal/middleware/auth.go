package middleware

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := utils.ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_provider", claims.Provider)
		c.Next()
	}
}

func CookieAuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwtService.ValidateAccessTokenFromCookie(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_provider", claims.Provider)
		c.Next()
	}
}

func SessionAuthMiddleware(sessionService *service.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := sessionService.ValidateSession(c)
		if err != nil || !result.IsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		c.Set("user_id", result.Claims.UserID)
		c.Set("user_email", result.Claims.Email)
		c.Set("user_name", result.Claims.Name)
		c.Set("user_provider", result.Claims.Provider)
		c.Set("session_id", result.Session.ID)
		c.Next()
	}
}

func FlexibleAuthMiddleware(jwtService *service.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwtService.ValidateAccessTokenFromCookie(c)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_name", claims.Name)
			c.Set("user_provider", claims.Provider)
			c.Set("auth_method", "cookie")
			c.Next()
			return
		}

		token, err := utils.ExtractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err = jwtService.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_name", claims.Name)
		c.Set("user_provider", claims.Provider)
		c.Set("auth_method", "bearer")
		c.Next()
	}
}
