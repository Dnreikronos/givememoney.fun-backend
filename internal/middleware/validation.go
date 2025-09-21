package middleware

import (
	"net/http"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Set("validator", validate)
		c.Next()
	})
}

func ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		validationErrors := make(map[string]string)

		if validatorErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validatorErrors {
				validationErrors[e.Field()] = getValidationErrorMessage(e)
			}
		} else {
			validationErrors["json"] = "Invalid JSON format"
		}

		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error:  "Validation failed",
			Fields: validationErrors,
		})
		return false
	}
	return true
}

func ValidateQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		validationErrors := make(map[string]string)

		if validatorErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validatorErrors {
				validationErrors[e.Field()] = getValidationErrorMessage(e)
			}
		} else {
			validationErrors["query"] = "Invalid query parameters"
		}

		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error:  "Validation failed",
			Fields: validationErrors,
		})
		return false
	}
	return true
}

func getValidationErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + e.Param() + " characters long"
	case "max":
		return "Must be at most " + e.Param() + " characters long"
	case "len":
		return "Must be exactly " + e.Param() + " characters long"
	case "gte":
		return "Must be greater than or equal to " + e.Param()
	case "lte":
		return "Must be less than or equal to " + e.Param()
	case "oneof":
		return "Must be one of: " + e.Param()
	default:
		return "Invalid value"
	}
}