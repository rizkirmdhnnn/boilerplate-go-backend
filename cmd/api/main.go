package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"boilerplate/internal/config"
	"boilerplate/internal/handler"
	"boilerplate/internal/middleware"
	"boilerplate/internal/repository"
	"boilerplate/internal/router"
	"boilerplate/internal/service"
	"boilerplate/pkg/database"

	"github.com/rs/zerolog/log"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	// Setup logging
	middleware.LogLevelSetter(cfg.LogLevel)
	middleware.LogFormatSetter(cfg.LogJSON)

	log.Info().
		Str("app", cfg.AppName).
		Str("version", cfg.AppVersion).
		Str("env", cfg.Env).
		Msg("starting server")

	// Database connection
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

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	userHdr := handler.NewUserHandler(userSvc)
	healthHdr := handler.NewHealthHandler(db)

	// Setup router
	r := router.Setup(&router.Config{
		JWTSecret:     cfg.JWTSecret,
		CORSOrigins:   cfg.CORSAllowedOrigins,
		DBPool:        db,
		AppVersion:    cfg.AppVersion,
		Debug:         cfg.Debug,
		UserHandler:   userHdr,
		HealthHandler: healthHdr,
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

	// Wait for signal
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited gracefully")
}
