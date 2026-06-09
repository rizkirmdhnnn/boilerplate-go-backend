package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	// Clear env vars that might interfere
	os.Clearenv()

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "boilerplate", cfg.AppName)
	assert.Equal(t, "1.0.0", cfg.AppVersion)
	assert.Equal(t, "development", cfg.Env)
	assert.False(t, cfg.Debug)
	assert.Equal(t, "0.0.0.0", cfg.ServerHost)
	assert.Equal(t, 8080, cfg.ServerPort)
	assert.Equal(t, "0.0.0.0:8080", cfg.ServerURL)
	assert.Equal(t, 10*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 10*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, "postgres", cfg.DBDriver)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, 5432, cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "postgres", cfg.DBPassword)
	assert.Equal(t, "boilerplate", cfg.DBName)
	assert.Equal(t, 25, cfg.DBMaxOpenConns)
	assert.Equal(t, 5, cfg.DBMaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.DBConnMaxLifetime)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, "change-me-in-production", cfg.JWTSecret)
	assert.Equal(t, 24*time.Hour, cfg.JWTExpiration)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.False(t, cfg.LogJSON)
}

func TestLoadFromEnv(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_NAME", "my-api")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("DB_HOST", "10.0.0.1")
	os.Setenv("DB_PASSWORD", "supersecret")
	os.Setenv("JWT_SECRET", "my-secret")
	os.Setenv("DEBUG", "true")
	os.Setenv("LOG_JSON", "true")
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://app1.com,https://app2.com")
	os.Setenv("DB_CONN_MAX_LIFETIME", "30s")

	cfg, err := Load("")
	require.NoError(t, err)

	assert.Equal(t, "my-api", cfg.AppName)
	assert.Equal(t, 9090, cfg.ServerPort)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "10.0.0.1", cfg.DBHost)
	assert.Equal(t, "supersecret", cfg.DBPassword)
	assert.Equal(t, "my-secret", cfg.JWTSecret)
	assert.True(t, cfg.LogJSON)
	assert.Equal(t, 30*time.Second, cfg.DBConnMaxLifetime)
	assert.Equal(t, []string{"https://app1.com", "https://app2.com"}, cfg.CORSAllowedOrigins)
}

func TestLoadInvalidDuration(t *testing.T) {
	os.Clearenv()
	os.Setenv("READ_TIMEOUT", "not-a-duration")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.ReadTimeout) // falls back to default
}

func TestLoadInvalidInt(t *testing.T) {
	os.Clearenv()
	os.Setenv("SERVER_PORT", "not-a-number")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.ServerPort) // falls back to default
}

func TestLoadCORSStar(t *testing.T) {
	os.Clearenv()
	os.Setenv("CORS_ALLOWED_ORIGINS", "*")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
}

func TestLoadEmptyDB(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_HOST", "")

	cfg, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, "", cfg.DBHost)
}
