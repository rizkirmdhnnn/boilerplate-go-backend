// @title           Boilerplate API
// @version         1.0.0
// @description     Go backend template with Clean Architecture
// @contact.name    Rizkirmdhn
// @license.name    MIT
// @host            localhost:8080
// @BasePath        /api/v1
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                Type "Bearer {token}" in the value field
package main

//go:generate echo "use 'make swag' or run 'swag init -g cmd/api/main.go --output docs' from project root"

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/application"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/config"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/middleware"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/repository"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/router"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/cache"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/database"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/migrator"

	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Logging
	middleware.LogLevelSetter(cfg.LogLevel)
	middleware.LogFormatSetter(cfg.LogJSON)

	log.Info().
		Str("app", cfg.AppName).
		Str("version", cfg.AppVersion).
		Str("env", cfg.Env).
		Msg("starting server")

	// Database — infrastructure adapter
	var db *database.Pool
	if cfg.DBHost != "" {
		dsn := database.DSN(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
		db, err = database.NewPool(dsn, cfg.DBMaxOpenConns, cfg.DBMaxIdleConns, cfg.DBConnMaxLifetime)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to database")
		}
		defer db.Close()
	} else {
		log.Warn().Msg("no database configured, running without DB")
	}

	// Auto-run database migrations
	if db != nil {
		if err := migrator.Run(cfg.DBMigrateDSN, "migrations"); err != nil {
			log.Fatal().Err(err).Msg("failed to run database migrations")
		}
	}

	// Dependency injection — wire ports to adapters
	userRepo := repository.NewUserRepository(db)                    // port → postgres adapter
	userSvc := application.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiration) // use-case

	// Cache — Redis or noop fallback
	var cacheClient cache.Cache
	if cfg.RedisEnabled {
		rc, err := cache.NewRedis(cache.Config{
			Host:     cfg.RedisHost,
			Port:     cfg.RedisPort,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to Redis")
		}
		cacheClient = rc
		log.Info().Msg("cache: using Redis")
	} else {
		cacheClient = cache.NewNoop()
		log.Info().Msg("cache: disabled (noop)")
	}
	defer cacheClient.Close()

	// Router
	r := router.Setup(&router.Config{
		JWTSecret:   cfg.JWTSecret,
		CORSOrigins: cfg.CORSAllowedOrigins,
		AppVersion:  cfg.AppVersion,
		Debug:       cfg.Debug,
		Cache:       cacheClient,
		UserSvc:     userSvc,
		DBPool:      db,

		RateLimitEnabled:        cfg.RateLimitEnabled,
		RateLimitRequestsPerMin: cfg.RateLimitRequestsPerMin,
		RateLimitBurst:          cfg.RateLimitBurst,
	})

	// HTTP server
	srv := &http.Server{
		Addr:         cfg.ServerURL,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("addr", cfg.ServerURL).Msg("listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited gracefully")
}
