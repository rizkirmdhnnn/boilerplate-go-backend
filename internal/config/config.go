package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppVersion string
	Env        string
	Debug      bool

	ServerHost string
	ServerPort int
	ServerURL  string

	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	DBDriver          string
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBMigrateDSN      string // computed: full DSN for migrations

	CORSAllowedOrigins []string

	JWTSecret     string
	JWTExpiration time.Duration

	LogLevel string
	LogJSON  bool
}

func Load(envPath ...string) (*Config, error) {
	// Load .env if exists
	envFile := ".env"
	if len(envPath) > 0 && envPath[0] != "" {
		envFile = envPath[0]
	}

	_ = godotenv.Load(envFile) // ignore error if .env doesn't exist

	cfg := &Config{
		AppName:    getEnv("APP_NAME", "boilerplate"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		Env:        getEnv("ENV", "development"),
		Debug:      getEnvBool("DEBUG", false),

		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvInt("SERVER_PORT", 8080),

		ReadTimeout:     getEnvDuration("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),

		DBDriver:          getEnv("DB_DRIVER", "postgres"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "boilerplate"),
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		JWTSecret:     getEnv("JWT_SECRET", "change-me-in-production"),
		JWTExpiration: getEnvDuration("JWT_EXPIRATION", 24*time.Hour),

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogJSON:  getEnvBool("LOG_JSON", false),
	}

	cfg.ServerURL = fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)

	// Build full DSN for migrations
	pw := cfg.DBPassword
	migDSN := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%d", cfg.DBHost, cfg.DBPort),
		User:   url.UserPassword(cfg.DBUser, pw),
		Path:   cfg.DBName,
	}
	migDSN.RawQuery = "sslmode=disable"
	cfg.DBMigrateDSN = migDSN.String()

	origins := getEnv("CORS_ALLOWED_ORIGINS", "*")
	if origins == "*" {
		cfg.CORSAllowedOrigins = []string{"*"}
	} else {
		cfg.CORSAllowedOrigins = splitAndTrim(origins, ",")
	}

	return cfg, nil
}

// helpers

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range split(s, sep) {
		part = trim(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// simple split without importing strings in config
func split(s, sep string) []string {
	var result []string
	var current []byte
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, string(current))
			current = nil
			i += len(sep) - 1
		} else {
			current = append(current, s[i])
		}
	}
	result = append(result, string(current))
	return result
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
