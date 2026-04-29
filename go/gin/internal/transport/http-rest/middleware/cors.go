package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed, all := false, false

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
			}
			if allowedOrigin == "*" {
				all = true
			}
		}

		if all {
			c.Header("Access-Control-Allow-Headers",
				"Content-Type,Authorization,Token,Set-Cookie,Cache-Control,Connection,Accept,Origin,Last-Event-ID,X-Requested-With,X-CSRF-Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, PATCH, HEAD, OPTIONS")
		}

		if allowed {
			c.Header("Access-Control-Allow-Headers",
				"Content-Type,Authorization,Token,Set-Cookie,Cache-Control,Connection,Accept,Origin,Last-Event-ID,X-Requested-With,X-CSRF-Token")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, PATCH, HEAD, OPTIONS")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
