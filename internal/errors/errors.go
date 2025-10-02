package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents application-specific error codes
type ErrorCode string

const (
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeTokenInvalid       ErrorCode = "TOKEN_INVALID"
	ErrorCodeUnauthorized       ErrorCode = "UNAUTHORIZED"

	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"

	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrorCodeConflict      ErrorCode = "CONFLICT"

	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"

	ErrorCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	ErrorCodeExternalServiceError ErrorCode = "EXTERNAL_SERVICE_ERROR"
	ErrorCodeProviderAPIError     ErrorCode = "PROVIDER_API_ERROR"
)

type AppError struct {
	Code       ErrorCode
	Message    string
	Err        error
	StatusCode int
	Context    map[string]interface{}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		StatusCode: getDefaultStatusCode(code),
		Context:    make(map[string]interface{}),
	}
}

func NewValidationError(message string, err error) *AppError {
	return NewAppError(ErrorCodeValidationFailed, message, err)
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), nil)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError(ErrorCodeUnauthorized, message, nil)
}

func NewDatabaseError(operation string, err error) *AppError {
	return NewAppError(ErrorCodeDatabaseError, fmt.Sprintf("Database operation failed: %s", operation), err)
}

func NewInternalError(message string, err error) *AppError {
	return NewAppError(ErrorCodeInternalError, message, err)
}

func NewProviderAPIError(message string, err error) *AppError {
	return NewAppError(ErrorCodeProviderAPIError, message, err)
}

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
	case ErrorCodeDatabaseError, ErrorCodeInternalError, ErrorCodeExternalServiceError, ErrorCodeProviderAPIError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

func WrapError(err error, code ErrorCode, message string) *AppError {
	return NewAppError(code, message, err)
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    ErrorCode              `json:"code"`
	Context map[string]interface{} `json:"context,omitempty"`
}

func (e *AppError) ToErrorResponse() ErrorResponse {
	return ErrorResponse{
		Error:   "error",
		Message: e.Message,
		Code:    e.Code,
		Context: e.Context,
	}
}
