package tests

import (
	"os"
	"testing"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type JWTServiceTestSuite struct {
	suite.Suite
	jwtService *service.JWTService
}

func (suite *JWTServiceTestSuite) SetupTest() {
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-long-enough-for-security")
	os.Setenv("GO_ENV", "test")

	suite.jwtService = service.NewJWTService()
}

func (suite *JWTServiceTestSuite) TearDownTest() {
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("GO_ENV")
}

func (suite *JWTServiceTestSuite) TestGenerateToken() {
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"
	provider := "twitch"

	token, err := suite.jwtService.GenerateToken(userID, email, name, provider)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), token)
}

func (suite *JWTServiceTestSuite) TestValidateToken() {
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"
	provider := "twitch"

	token, err := suite.jwtService.GenerateToken(userID, email, name, provider)
	assert.NoError(suite.T(), err)

	claims, err := suite.jwtService.ValidateToken(token)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, claims.UserID)
	assert.Equal(suite.T(), email, claims.Email)
	assert.Equal(suite.T(), name, claims.Name)
	assert.Equal(suite.T(), provider, claims.Provider)
}

func (suite *JWTServiceTestSuite) TestValidateInvalidToken() {
	invalidToken := "invalid.token.here"

	claims, err := suite.jwtService.ValidateToken(invalidToken)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
}

func (suite *JWTServiceTestSuite) TestGenerateTokenPair() {
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"
	provider := "twitch"

	tokenPair, err := suite.jwtService.GenerateTokenPair(userID, email, name, provider)

	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), tokenPair.AccessToken)
	assert.NotEmpty(suite.T(), tokenPair.RefreshToken)
	assert.Equal(suite.T(), int64(900), tokenPair.ExpiresIn) // 15 minutes

	accessClaims, err := suite.jwtService.ValidateAccessToken(tokenPair.AccessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "access", accessClaims.TokenType)

	refreshClaims, err := suite.jwtService.ValidateToken(tokenPair.RefreshToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "refresh", refreshClaims.TokenType)
}

func (suite *JWTServiceTestSuite) TestRefreshToken() {
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"
	provider := "twitch"

	initialPair, err := suite.jwtService.GenerateTokenPair(userID, email, name, provider)
	assert.NoError(suite.T(), err)

	newPair, err := suite.jwtService.RefreshToken(initialPair.RefreshToken)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), newPair.AccessToken)
	assert.NotEmpty(suite.T(), newPair.RefreshToken)

	accessClaims, err := suite.jwtService.ValidateAccessToken(newPair.AccessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), userID, accessClaims.UserID)

	refreshClaims, err := suite.jwtService.ValidateToken(newPair.RefreshToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "refresh", refreshClaims.TokenType)
}

func (suite *JWTServiceTestSuite) TestValidateAccessTokenRejectsRefreshToken() {
	userID := uuid.New()
	email := "test@example.com"
	name := "Test User"
	provider := "twitch"

	tokenPair, err := suite.jwtService.GenerateTokenPair(userID, email, name, provider)
	assert.NoError(suite.T(), err)

	claims, err := suite.jwtService.ValidateAccessToken(tokenPair.RefreshToken)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), claims)
	assert.Contains(suite.T(), err.Error(), "invalid token type")
}

func TestJWTServiceTestSuite(t *testing.T) {
	suite.Run(t, new(JWTServiceTestSuite))
}
