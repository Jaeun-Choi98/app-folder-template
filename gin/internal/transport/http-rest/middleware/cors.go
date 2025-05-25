package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewCORSMiddleware(origins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", strings.Join(origins, ","))
		c.Writer.Header().Add("Access-Control-Allow-Headers",
			"Content-Type,AccessToken,X-CSRF-Token,Authorization,Token,Set-Cookie,X-Requested-With")
		c.Writer.Header().Add("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		//w.Header().Set("content-type", "application/json;charset=UTF-8")
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}
		c.Next()
	}
}
