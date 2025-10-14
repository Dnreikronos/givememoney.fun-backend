package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/dto"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	cleanup  time.Duration
}

func NewIPRateLimiter(rps rate.Limit, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rps,
		burst:    burst,
		cleanup:  time.Minute * 10,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

func (i *IPRateLimiter) CleanupOldLimiters() {
	ticker := time.NewTicker(i.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.mu.Lock()
			for ip, limiter := range i.limiters {
				if limiter.Tokens() == float64(i.burst) {
					delete(i.limiters, ip)
				}
			}
			i.mu.Unlock()
		}
	}
}

var globalRateLimiter *IPRateLimiter

func InitializeRateLimiter(requestsPerSecond float64, burst int) {
	globalRateLimiter = NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)
	go globalRateLimiter.CleanupOldLimiters()
}

func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	if globalRateLimiter == nil {
		InitializeRateLimiter(requestsPerSecond, burst)
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := globalRateLimiter.GetLimiter(ip)

		if !limiter.Allow() {
			reservation := limiter.Reserve()
			resetTime := time.Now().Add(reservation.Delay())
			reservation.Cancel()

			c.Header("X-RateLimit-Limit", strconv.Itoa(burst))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))
			c.Header("Retry-After", strconv.FormatInt(int64(reservation.Delay().Seconds()), 10))

			c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error:   "Rate limit exceeded",
				Message: "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		remaining := int(limiter.Tokens())
		if remaining > burst {
			remaining = burst
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(burst))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		c.Next()
	}
}

func StrictRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(5, 10)
}

func AuthRateLimitMiddleware() gin.HandlerFunc {
	return RateLimitMiddleware(2, 5)
}
