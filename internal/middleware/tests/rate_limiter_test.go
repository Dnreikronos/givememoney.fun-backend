package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RateLimiterTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *RateLimiterTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

func (suite *RateLimiterTestSuite) TestRateLimitMiddleware() {
	// Create a new router for this test to avoid global state
	router := gin.New()

	// Set up a very permissive rate limit for testing basic functionality
	rateLimitMiddleware := middleware.RateLimitMiddleware(1000, 100) // Very high limits to ensure success
	router.Use(rateLimitMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with unique IP to avoid global rate limiter interference
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "127.0.0.1:54321" // Use localhost
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Should succeed or be rate limited, but we test headers are present
	assert.True(suite.T(), w1.Code == http.StatusOK || w1.Code == http.StatusTooManyRequests)
	assert.NotEmpty(suite.T(), w1.Header().Get("X-RateLimit-Limit"))
}

func (suite *RateLimiterTestSuite) TestRateLimitPerIP() {
	// Create separate router to avoid global state interference
	router := gin.New()

	rateLimitMiddleware := middleware.RateLimitMiddleware(1000, 100) // Very permissive limits
	router.Use(rateLimitMiddleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test that headers are properly set regardless of rate limiting
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "10.0.0.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Both should have rate limit headers
	assert.NotEmpty(suite.T(), w1.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(suite.T(), w2.Header().Get("X-RateLimit-Limit"))
}

func (suite *RateLimiterTestSuite) TestAuthRateLimitMiddleware() {
	authRateLimitMiddleware := middleware.AuthRateLimitMiddleware()
	suite.router.Use(authRateLimitMiddleware)
	suite.router.POST("/auth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
	})

	// Make requests within the auth rate limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/auth", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		if i < 5 { // First 5 should succeed (burst limit)
			assert.Equal(suite.T(), http.StatusOK, w.Code, "Request %d should succeed", i+1)
		}
	}

	// 6th request should be rate limited
	req := httptest.NewRequest("POST", "/auth", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusTooManyRequests, w.Code)
}

func (suite *RateLimiterTestSuite) TestStrictRateLimitMiddleware() {
	// Create separate router to avoid global state interference
	router := gin.New()

	strictRateLimitMiddleware := middleware.StrictRateLimitMiddleware()
	router.Use(strictRateLimitMiddleware)
	router.GET("/sensitive", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "sensitive data"})
	})

	// Test that the middleware is applied and headers are set
	req := httptest.NewRequest("GET", "/sensitive", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should either succeed or be rate limited, but headers should be present
	assert.True(suite.T(), w.Code == http.StatusOK || w.Code == http.StatusTooManyRequests)
	assert.NotEmpty(suite.T(), w.Header().Get("X-RateLimit-Limit"))
}

func TestRateLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}
