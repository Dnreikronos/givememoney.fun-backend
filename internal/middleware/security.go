package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Server", "")

		// The alert overlay page is embedded in OBS Browser Source (iframe)
		// and uses inline styles/scripts + Google Fonts, so it needs relaxed CSP.
		if strings.HasPrefix(c.Request.URL.Path, "/api/alerts/") {
			c.Header("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'unsafe-inline'; "+
					"style-src 'unsafe-inline' https://fonts.googleapis.com; "+
					"font-src https://fonts.gstatic.com; "+
					"connect-src 'self' ws: wss:; "+
					"object-src 'none';")
		} else {
			c.Header("X-Frame-Options", "DENY")
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")
		}

		if c.Request.Header.Get("X-Forwarded-Proto") == "https" || c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
