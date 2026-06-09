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

//go:generate swag init -g main.go --output ../docs --quiet

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"boilerplate/internal/application"
	"boilerplate/internal/config"
	"boilerplate/internal/middleware"
	"boilerplate/internal/repository"
	"boilerplate/internal/router"
	"boilerplate/pkg/database"

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

	// Dependency injection — wire ports to adapters
	userRepo := repository.NewUserRepository(db)                    // port → postgres adapter
	userSvc := application.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiration) // use-case

	// Router
	r := router.Setup(&router.Config{
		JWTSecret:   cfg.JWTSecret,
		CORSOrigins: cfg.CORSAllowedOrigins,
		AppVersion:  cfg.AppVersion,
		Debug:       cfg.Debug,
		UserSvc:     userSvc,
		DBPool:      db,
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
