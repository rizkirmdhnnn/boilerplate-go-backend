package middleware

import (
	"net/http"
	"strings"

	"boilerplate/internal/service"
	"boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthRequired validates JWT tokens from the Authorization header.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := service.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// CORS returns a Gin middleware that handles CORS headers.
// Note: This is already defined in cors.go — keeping this as an alias.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return CORS(allowedOrigins)
}
