package middleware

import (
	"net/http"
	"strings"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/application"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthRequired validates JWT tokens from the Authorization header.
// Uses application.ValidateJWT so the auth logic lives in the use-case layer.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := application.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}
