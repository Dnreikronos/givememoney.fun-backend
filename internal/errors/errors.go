package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	// Authentication errors
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrorCodeUnauthorized       ErrorCode = "UNAUTHORIZED"

	// Validation errors
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"

	// Resource errors
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrorCodeConflict      ErrorCode = "CONFLICT"

	// System errors
	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	// Rate limiting
	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// External service errors
	ErrorCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeTwitchAPIError       ErrorCode = "TWITCH_API_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode
	Message    string
	Err        error
	StatusCode int
	Context    map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		StatusCode: getDefaultStatusCode(code),
		Context:    make(map[string]interface{}),
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string, err error) *AppError {
	return NewAppError(ErrorCodeValidationFailed, message, err)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), nil)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrorCodeUnauthorized, message, nil)
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation string, err error) *AppError {
	return NewAppError(ErrorCodeDatabaseError, fmt.Sprintf("Database operation failed: %s", operation), err)
}

// NewInternalError creates an internal server error
func NewInternalError(message string, err error) *AppError {
	return NewAppError(ErrorCodeInternalError, message, err)
}

// NewTwitchAPIError creates a Twitch API error
func NewTwitchAPIError(message string, err error) *AppError {
	return NewAppError(ErrorCodeTwitchAPIError, message, err)
}

// getDefaultStatusCode returns the default HTTP status code for an error code
func getDefaultStatusCode(code ErrorCode) int {
	switch code {
	case ErrorCodeInvalidCredentials, ErrorCodeTokenExpired, ErrorCodeTokenInvalid, ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrorCodeValidationFailed, ErrorCodeInvalidInput:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeAlreadyExists, ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeDatabaseError, ErrorCodeInternalError, ErrorCodeExternalServiceError, ErrorCodeTwitchAPIError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps an existing error with additional context
func WrapError(err error, code ErrorCode, message string) *AppError {
	return NewAppError(code, message, err)
}

// ErrorResponse represents an error response for the API
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    ErrorCode              `json:"code"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ToErrorResponse converts an AppError to an ErrorResponse
func (e *AppError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error:   "error",
		Message: e.Message,
		Code:    e.Code,
		Context: e.Context,
	}
}
