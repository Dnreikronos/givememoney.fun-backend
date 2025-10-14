package middleware

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func ValidateRegisterRequest(c *gin.Context, req *dto.RegisterRequest) bool {
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: "Validation failed",
			Fields: map[string]string{
				"password": err.Error(),
			},
		})
		return false
	}

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
