package router

import (
	_ "boilerplate/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"boilerplate/internal/application"
	"boilerplate/internal/handler"
	"boilerplate/internal/middleware"
	"boilerplate/pkg/database"

	"github.com/gin-gonic/gin"
)

// Config holds router dependencies.
// All fields are interfaces (ports) — actual implementations are injected in main.
type Config struct {
	JWTSecret     string
	CORSOrigins   []string
	AppVersion    string
	Debug         bool
	UserSvc       application.UserService
	DBPool        *database.Pool

	// Rate limiter settings
	RateLimitEnabled       bool
	RateLimitRequestsPerMin int
	RateLimitBurst         int
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
	r.Use(middleware.ErrorHandler())

	// Rate limiter (conditional)
	if cfg.RateLimitEnabled {
		rl := middleware.NewRateLimiterWithParams(cfg.RateLimitRequestsPerMin, cfg.RateLimitBurst)
		r.Use(rl.RateLimit())
	}

	r.Use(func(c *gin.Context) {
		c.Set("app_version", cfg.AppVersion)
		c.Next()
	})

	// Handlers
	healthHdr := handler.NewHealthHandler(cfg.DBPool)
	userHdr := handler.NewUserHandler(cfg.UserSvc)

	// Health — no auth
	r.GET("/health", healthHdr.Check)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := r.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userHdr.Register)
		auth.POST("/login", userHdr.Login)
	}

	// Protected
	protected := v1.Group("")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		users := protected.Group("/users")
		{
			users.GET("/me", userHdr.GetProfile)
			users.GET("", userHdr.ListUsers)
			users.GET("/:id", userHdr.GetUser)
			users.PUT("/:id", userHdr.UpdateUser)
			users.DELETE("/:id", userHdr.DeleteUser)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success": false,
			"error":   gin.H{"message": "route not found"},
		})
	})

	return r
}
