package tests

import (
	"errors"
	"net/http"
	"testing"

	appErrors "github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ErrorsTestSuite struct {
	suite.Suite
}

func (suite *ErrorsTestSuite) TestNewAppError() {
	originalErr := errors.New("original error")
	appErr := appErrors.NewAppError(appErrors.ErrorCodeDatabaseError, "Database operation failed", originalErr)

	assert.Equal(suite.T(), appErrors.ErrorCodeDatabaseError, appErr.Code)
	assert.Equal(suite.T(), "Database operation failed", appErr.Message)
	assert.Equal(suite.T(), originalErr, appErr.Err)
	assert.Equal(suite.T(), http.StatusInternalServerError, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestErrorString() {
	originalErr := errors.New("original error")
	appErr := appErrors.NewAppError(appErrors.ErrorCodeDatabaseError, "Database operation failed", originalErr)

	expectedError := "DATABASE_ERROR: Database operation failed (caused by: original error)"
	assert.Equal(suite.T(), expectedError, appErr.Error())
}

func (suite *ErrorsTestSuite) TestErrorStringWithoutOriginalError() {
	appErr := appErrors.NewAppError(appErrors.ErrorCodeNotFound, "User not found", nil)

	expectedError := "NOT_FOUND: User not found"
	assert.Equal(suite.T(), expectedError, appErr.Error())
}

func (suite *ErrorsTestSuite) TestUnwrapError() {
	originalErr := errors.New("original error")
	appErr := appErrors.NewAppError(appErrors.ErrorCodeDatabaseError, "Database operation failed", originalErr)

	unwrapped := appErr.Unwrap()
	assert.Equal(suite.T(), originalErr, unwrapped)
}

func (suite *ErrorsTestSuite) TestWithContext() {
	appErr := appErrors.NewAppError(appErrors.ErrorCodeValidationFailed, "Validation failed", nil)
	appErr.WithContext("field", "email")
	appErr.WithContext("value", "invalid-email")

	assert.Equal(suite.T(), "email", appErr.Context["field"])
	assert.Equal(suite.T(), "invalid-email", appErr.Context["value"])
}

func (suite *ErrorsTestSuite) TestNewValidationError() {
	originalErr := errors.New("validation error")
	appErr := appErrors.NewValidationError("Invalid input", originalErr)

	assert.Equal(suite.T(), appErrors.ErrorCodeValidationFailed, appErr.Code)
	assert.Equal(suite.T(), "Invalid input", appErr.Message)
	assert.Equal(suite.T(), originalErr, appErr.Err)
	assert.Equal(suite.T(), http.StatusBadRequest, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestNewNotFoundError() {
	appErr := appErrors.NewNotFoundError("user")

	assert.Equal(suite.T(), appErrors.ErrorCodeNotFound, appErr.Code)
	assert.Equal(suite.T(), "user not found", appErr.Message)
	assert.Nil(suite.T(), appErr.Err)
	assert.Equal(suite.T(), http.StatusNotFound, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestNewUnauthorizedError() {
	appErr := appErrors.NewUnauthorizedError("Invalid token")

	assert.Equal(suite.T(), appErrors.ErrorCodeUnauthorized, appErr.Code)
	assert.Equal(suite.T(), "Invalid token", appErr.Message)
	assert.Nil(suite.T(), appErr.Err)
	assert.Equal(suite.T(), http.StatusUnauthorized, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestNewDatabaseError() {
	originalErr := errors.New("connection failed")
	appErr := appErrors.NewDatabaseError("select operation", originalErr)

	assert.Equal(suite.T(), appErrors.ErrorCodeDatabaseError, appErr.Code)
	assert.Equal(suite.T(), "Database operation failed: select operation", appErr.Message)
	assert.Equal(suite.T(), originalErr, appErr.Err)
	assert.Equal(suite.T(), http.StatusInternalServerError, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestNewTwitchAPIError() {
	originalErr := errors.New("API rate limit exceeded")
	appErr := appErrors.NewTwitchAPIError("Failed to get user info", originalErr)

	assert.Equal(suite.T(), appErrors.ErrorCodeTwitchAPIError, appErr.Code)
	assert.Equal(suite.T(), "Failed to get user info", appErr.Message)
	assert.Equal(suite.T(), originalErr, appErr.Err)
	assert.Equal(suite.T(), http.StatusInternalServerError, appErr.StatusCode)
}

func (suite *ErrorsTestSuite) TestIsAppError() {
	// Test with AppError
	appErr := appErrors.NewNotFoundError("user")
	detectedErr, isAppErr := appErrors.IsAppError(appErr)
	assert.True(suite.T(), isAppErr)
	assert.Equal(suite.T(), appErr, detectedErr)

	// Test with regular error
	regularErr := errors.New("regular error")
	detectedErr, isAppErr = appErrors.IsAppError(regularErr)
	assert.False(suite.T(), isAppErr)
	assert.Nil(suite.T(), detectedErr)
}

func (suite *ErrorsTestSuite) TestWrapError() {
	originalErr := errors.New("original error")
	wrappedErr := appErrors.WrapError(originalErr, appErrors.ErrorCodeExternalServiceError, "Service failed")

	assert.Equal(suite.T(), appErrors.ErrorCodeExternalServiceError, wrappedErr.Code)
	assert.Equal(suite.T(), "Service failed", wrappedErr.Message)
	assert.Equal(suite.T(), originalErr, wrappedErr.Err)
}

func (suite *ErrorsTestSuite) TestToErrorResponse() {
	appErr := appErrors.NewValidationError("Invalid email format", nil)
	appErr.WithContext("field", "email")

	response := appErr.ToErrorResponse()

	assert.Equal(suite.T(), "error", response.Error)
	assert.Equal(suite.T(), "Invalid email format", response.Message)
	assert.Equal(suite.T(), appErrors.ErrorCodeValidationFailed, response.Code)
	assert.Equal(suite.T(), "email", response.Context["field"])
}

func (suite *ErrorsTestSuite) TestDefaultStatusCodes() {
	testCases := []struct {
		code           appErrors.ErrorCode
		expectedStatus int
	}{
		{appErrors.ErrorCodeInvalidCredentials, http.StatusUnauthorized},
		{appErrors.ErrorCodeTokenExpired, http.StatusUnauthorized},
		{appErrors.ErrorCodeUnauthorized, http.StatusUnauthorized},
		{appErrors.ErrorCodeValidationFailed, http.StatusBadRequest},
		{appErrors.ErrorCodeInvalidInput, http.StatusBadRequest},
		{appErrors.ErrorCodeNotFound, http.StatusNotFound},
		{appErrors.ErrorCodeAlreadyExists, http.StatusConflict},
		{appErrors.ErrorCodeConflict, http.StatusConflict},
		{appErrors.ErrorCodeRateLimitExceeded, http.StatusTooManyRequests},
		{appErrors.ErrorCodeServiceUnavailable, http.StatusServiceUnavailable},
		{appErrors.ErrorCodeDatabaseError, http.StatusInternalServerError},
		{appErrors.ErrorCodeInternalError, http.StatusInternalServerError},
		{appErrors.ErrorCodeExternalServiceError, http.StatusInternalServerError},
		{appErrors.ErrorCodeTwitchAPIError, http.StatusInternalServerError},
	}

	for _, tc := range testCases {
		appErr := appErrors.NewAppError(tc.code, "Test message", nil)
		assert.Equal(suite.T(), tc.expectedStatus, appErr.StatusCode, "Error code %s should have status %d", tc.code, tc.expectedStatus)
	}
}

func TestErrorsTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorsTestSuite))
}