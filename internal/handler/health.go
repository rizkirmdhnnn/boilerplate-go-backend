package handler

import (
	"net/http"

	"boilerplate/pkg/database"

	"github.com/gin-gonic/gin"
)

// HealthHandler holds dependencies for health checks.
type HealthHandler struct {
	db *database.Pool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *database.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthResponse is the payload for /health.
type HealthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
}

// Check responds with service health status.
func (h *HealthHandler) Check(c *gin.Context) {
	version := c.GetString("app_version")
	if version == "" {
		version = "unknown"
	}

	dbStatus := "up"
	if h.db != nil {
		if err := h.db.HealthCheck(c.Request.Context()); err != nil {
			dbStatus = "down"
		}
	}

	status := "ok"
	if dbStatus == "down" {
		status = "degraded"
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status:    status,
		Version:   version,
		Timestamp: c.GetString("request_time"),
		Database:  dbStatus,
	})
}
