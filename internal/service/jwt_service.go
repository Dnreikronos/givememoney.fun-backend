package service

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type CookieConfig struct {
	SecureCookies bool
	Domain        string
	SameSite      http.SameSite
}

type JWTService struct {
	secret       []byte
	cookieConfig CookieConfig
}

func NewJWTService() *JWTService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	if len(secret) < 32 {
		panic("JWT_SECRET must be at least 32 characters long")
	}

	cookieConfig := CookieConfig{
		SecureCookies: os.Getenv("GO_ENV") == "production",
		Domain:        os.Getenv("COOKIE_DOMAIN"), // Can be empty for same-origin
		SameSite:      http.SameSiteLaxMode,       // Default to Lax for better compatibility
	}

	if sameSite := os.Getenv("COOKIE_SAME_SITE"); sameSite != "" {
		switch sameSite {
		case "strict":
			cookieConfig.SameSite = http.SameSiteStrictMode
		case "none":
			cookieConfig.SameSite = http.SameSiteNoneMode
		default:
			cookieConfig.SameSite = http.SameSiteLaxMode
		}
	}

	return &JWTService{
		secret:       []byte(secret),
		cookieConfig: cookieConfig,
	}
}

func (j *JWTService) GenerateToken(userID uuid.UUID, email, name, provider string) (string, error) {
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		Provider:  provider,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "givememoney-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) GenerateTokenPair(userID uuid.UUID, email, name, provider string) (*TokenPair, error) {
	accessClaims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		Provider:  provider,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // Reduced from 1 hour
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "givememoney-backend",
			Audience:  []string{"givememoney-frontend"}, // Add audience for additional security
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secret)
	if err != nil {
		return nil, err
	}

	refreshClaims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Name:      name,
		Provider:  provider,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Reduced from 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "givememoney-backend",
			Audience:  []string{"givememoney-frontend"}, // Add audience for additional security
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    900, // 15 minutes in seconds
	}, nil
}

func (j *JWTService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	return j.GenerateTokenPair(claims.UserID, claims.Email, claims.Name, claims.Provider)
}

func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		expectedAudience := "givememoney-frontend"
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == expectedAudience {
				validAudience = true
				break
			}
		}
		if !validAudience && len(claims.Audience) > 0 {
			return nil, errors.New("invalid audience")
		}

		if claims.Issuer != "givememoney-backend" && claims.Issuer != "" {
			return nil, errors.New("invalid issuer")
		}

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

func (j *JWTService) SetTokenCookies(c *gin.Context, tokenPair *TokenPair) {
	c.SetSameSite(j.cookieConfig.SameSite)
	c.SetCookie(
		"auth_token",
		tokenPair.AccessToken,
		900,
		"/",
		j.cookieConfig.Domain,
		j.cookieConfig.SecureCookies,
		true,
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		24*3600,
		"/",
		j.cookieConfig.Domain,
		j.cookieConfig.SecureCookies,
		true,
	)

	c.SetCookie(
		"session_id",
		"session_active",
		900,
		"/",
		j.cookieConfig.Domain,
		j.cookieConfig.SecureCookies,
		false,
	)
}

func (j *JWTService) ClearTokenCookies(c *gin.Context) {
	cookies := []string{"auth_token", "refresh_token", "session_id"}
	for _, cookieName := range cookies {
		c.SetCookie(
			cookieName,
			"",
			-1,
			"/",
			j.cookieConfig.Domain,
			j.cookieConfig.SecureCookies,
			true,
		)
	}
}

func (j *JWTService) GetTokenFromCookie(c *gin.Context, tokenType string) (string, error) {
	var cookieName string
	switch tokenType {
	case "access":
		cookieName = "auth_token"
	case "refresh":
		cookieName = "refresh_token"
	default:
		return "", errors.New("invalid token type")
	}

	token, err := c.Cookie(cookieName)
	if err != nil {
		return "", errors.New("token not found in cookies")
	}

	return token, nil
}

func (j *JWTService) ValidateAccessTokenFromCookie(c *gin.Context) (*JWTClaims, error) {
	token, err := j.GetTokenFromCookie(c, "access")
	if err != nil {
		return nil, err
	}

	return j.ValidateAccessToken(token)
}

func (j *JWTService) RefreshTokenFromCookie(c *gin.Context) (*TokenPair, error) {
	refreshToken, err := j.GetTokenFromCookie(c, "refresh")
	if err != nil {
		return nil, err
	}

	newTokenPair, err := j.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	j.SetTokenCookies(c, newTokenPair)

	return newTokenPair, nil
}
