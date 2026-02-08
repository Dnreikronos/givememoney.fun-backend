package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type RequestLoggerTestSuite struct {
	suite.Suite
}

func (s *RequestLoggerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
}

func (s *RequestLoggerTestSuite) newObservedRouter(level zap.AtomicLevel) (*gin.Engine, *observer.ObservedLogs) {
	core, logs := observer.New(level.Level())
	logger := zap.New(core)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.RequestLoggerMiddleware(logger))
	return router, logs
}

func (s *RequestLoggerTestSuite) TestLogsInfoForSuccessfulRequest() {
	router, logs := s.newObservedRouter(zap.NewAtomicLevelAt(zap.DebugLevel))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(s.T(), zap.InfoLevel, entry.Level)
	assert.Equal(s.T(), "HTTP request", entry.Message)

	fields := fieldMap(entry.ContextMap())
	assert.Equal(s.T(), "GET", fields["method"])
	assert.Equal(s.T(), "/test", fields["path"])
	assert.Equal(s.T(), "foo=bar", fields["query"])
	assert.Equal(s.T(), int64(http.StatusOK), fields["status"])
	assert.Equal(s.T(), "test-agent", fields["user_agent"])
	assert.NotEmpty(s.T(), fields["request_id"])
}

func (s *RequestLoggerTestSuite) TestLogsWarnFor4xx() {
	router, logs := s.newObservedRouter(zap.NewAtomicLevelAt(zap.DebugLevel))
	router.GET("/notfound", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	req := httptest.NewRequest("GET", "/notfound", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), 1, logs.Len())
	assert.Equal(s.T(), zap.WarnLevel, logs.All()[0].Level)
}

func (s *RequestLoggerTestSuite) TestLogsErrorFor5xx() {
	router, logs := s.newObservedRouter(zap.NewAtomicLevelAt(zap.DebugLevel))
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), 1, logs.Len())
	assert.Equal(s.T(), zap.ErrorLevel, logs.All()[0].Level)
}

func (s *RequestLoggerTestSuite) TestIncludesStreamerID() {
	router, logs := s.newObservedRouter(zap.NewAtomicLevelAt(zap.DebugLevel))
	router.GET("/api/streamer/:streamer_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/api/streamer/abc123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), 1, logs.Len())
	fields := fieldMap(logs.All()[0].ContextMap())
	assert.Equal(s.T(), "abc123", fields["streamer_id"])
}

func (s *RequestLoggerTestSuite) TestOmitsStreamerIDWhenAbsent() {
	router, logs := s.newObservedRouter(zap.NewAtomicLevelAt(zap.DebugLevel))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(s.T(), 1, logs.Len())
	fields := fieldMap(logs.All()[0].ContextMap())
	_, exists := fields["streamer_id"]
	assert.False(s.T(), exists)
}

func fieldMap(m map[string]interface{}) map[string]interface{} {
	return m
}

func TestRequestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(RequestLoggerTestSuite))
}
