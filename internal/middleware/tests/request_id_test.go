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

type RequestIDTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (s *RequestIDTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.Use(middleware.RequestIDMiddleware())
	s.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": c.GetString(middleware.RequestIDKey)})
	})
}

func (s *RequestIDTestSuite) TestGeneratesUUIDWhenHeaderMissing() {
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)

	requestID := w.Header().Get(middleware.RequestIDHeader)
	assert.NotEmpty(s.T(), requestID)
	assert.Len(s.T(), requestID, 36) // UUID v4 format
}

func (s *RequestIDTestSuite) TestUsesExistingHeader() {
	existingID := "my-custom-request-id-123"

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(middleware.RequestIDHeader, existingID)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	assert.Equal(s.T(), http.StatusOK, w.Code)
	assert.Equal(s.T(), existingID, w.Header().Get(middleware.RequestIDHeader))
}

func (s *RequestIDTestSuite) TestSetsContextValue() {
	var contextValue string

	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.Use(middleware.RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		contextValue = c.GetString(middleware.RequestIDKey)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEmpty(s.T(), contextValue)
	assert.Equal(s.T(), w.Header().Get(middleware.RequestIDHeader), contextValue)
}

func (s *RequestIDTestSuite) TestUniqueIDsPerRequest() {
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	id1 := w1.Header().Get(middleware.RequestIDHeader)
	id2 := w2.Header().Get(middleware.RequestIDHeader)

	assert.NotEqual(s.T(), id1, id2)
}

func TestRequestIDTestSuite(t *testing.T) {
	suite.Run(t, new(RequestIDTestSuite))
}
