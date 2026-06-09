package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API response envelope.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// Meta holds pagination metadata.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// APIError holds error details.
type APIError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// OK sends a 200 success response.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Created sends a 201 response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// Message sends a 200 with just a message.
func Message(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: msg,
	})
}

// Paginated sends a paginated response.
func Paginated(c *gin.Context, data interface{}, page, perPage, total int) {
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error sends an error response with the given status code.
func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Message: msg,
		},
	})
}

// ErrorDetail sends an error with a code and details.
func ErrorDetail(c *gin.Context, status int, code, msg, details string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: msg,
			Details: details,
		},
	})
}

// ValidationError sends a 422 validation error.
func ValidationError(c *gin.Context, msg string) {
	if msg == "" {
		msg = "validation failed"
	}
	Error(c, http.StatusUnprocessableEntity, msg)
}

// NotFound sends a 404 error.
func NotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "resource not found"
	}
	Error(c, http.StatusNotFound, msg)
}

// InternalError sends a 500 error (hides details in production).
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "internal server error")
}

// Unauthorized sends a 401 error.
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "unauthorized"
	}
	Error(c, http.StatusUnauthorized, msg)
}

// Forbidden sends a 403 error.
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "forbidden"
	}
	Error(c, http.StatusForbidden, msg)
}
