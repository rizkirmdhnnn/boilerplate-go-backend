package router

import (
	"boilerplate/internal/handler"
	"boilerplate/internal/middleware"
	"boilerplate/pkg/database"

	"github.com/gin-gonic/gin"
)

// Config holds router configuration.
type Config struct {
	JWTSecret     string
	CORSOrigins   []string
	DBPool        *database.Pool
	AppVersion    string
	Debug         bool
	UserHandler   *handler.UserHandler
	HealthHandler *handler.HealthHandler
}

// Setup configures the Gin router with all routes and middleware.
func Setup(cfg *Config) *gin.Engine {
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.SecurityHeaders())

	// Set app version in context
	r.Use(func(c *gin.Context) {
		c.Set("app_version", cfg.AppVersion)
		c.Next()
	})

	// Health check — no auth
	r.GET("/health", cfg.HealthHandler.Check)

	// API v1
	v1 := r.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", cfg.UserHandler.Register)
		auth.POST("/login", cfg.UserHandler.Login)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		users := protected.Group("/users")
		{
			users.GET("/me", cfg.UserHandler.GetProfile)
			users.GET("", cfg.UserHandler.ListUsers)
			users.GET("/:id", cfg.UserHandler.GetUser)
			users.PUT("/:id", cfg.UserHandler.UpdateUser)
			users.DELETE("/:id", cfg.UserHandler.DeleteUser)
		}
	}

	// NoRoute handler
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error": gin.H{
				"message": "route not found",
			},
		})
	})

	return r
}
