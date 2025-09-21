package middleware

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

// ValidateRegisterRequest validates registration request including password confirmation
func ValidateRegisterRequest(c *gin.Context, req *dto.RegisterRequest) bool {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: "Validation failed",
			Fields: map[string]string{
				"password": err.Error(),
			},
		})
		return false
	}

	// Validate password confirmation
	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: "Validation failed",
			Fields: map[string]string{
				"confirm_password": "Passwords do not match",
			},
		})
		return false
	}

	return true
}

// ValidatePasswordStrengthMiddleware is a middleware that validates password strength in requests
func ValidatePasswordStrengthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This middleware will be applied before binding, so we'll let the controller handle validation
		c.Next()
	}
}