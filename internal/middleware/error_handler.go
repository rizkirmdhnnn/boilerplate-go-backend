package middleware

import (
	"errors"
	"net/http"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/application"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/domain"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a Gin middleware that catches errors added via c.Error()
// and maps them to proper HTTP status codes using errors.Is().
// It only acts on errors not already handled by the handler (i.e. when no
// response has been written yet).
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// If the handler already wrote a response, do not overwrite it.
		if c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err

		switch {
		case errors.Is(err, domain.ErrNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrDuplicate):
			response.Error(c, http.StatusConflict, err.Error())
		case errors.Is(err, application.ErrInvalidCredentials):
			response.Error(c, http.StatusUnauthorized, err.Error())
		case errors.Is(err, application.ErrAccountDisabled):
			response.Error(c, http.StatusUnauthorized, err.Error())
		default:
			response.InternalError(c)
		}
	}
}
