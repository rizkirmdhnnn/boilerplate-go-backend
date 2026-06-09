package middleware

import (
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics gracefully.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		response.InternalError(c)
		c.Abort()
	})
}

// SecurityHeaders adds security-related HTTP headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
