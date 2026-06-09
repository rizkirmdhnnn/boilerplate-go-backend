package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger returns a Gin middleware that logs requests using zerolog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Set request time in context
		c.Set("request_time", start.Format(time.RFC3339))

		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)

		event := log.Info()
		if c.Writer.Status() >= 500 {
			event = log.Error()
		} else if c.Writer.Status() >= 400 {
			event = log.Warn()
		}

		event.Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", c.Writer.Status()).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Int("body_size", c.Writer.Size()).
			Msg("request completed")

		// Log any errors attached by other handlers
		for _, err := range c.Errors {
			log.Error().Err(err.Err).Int("type", int(err.Type)).Msg("request error")
		}
	}
}

// LogLevelSetter configures zerolog level from config.
func LogLevelSetter(level string) {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
}

// LogFormatSetter configures zerolog output format.
func LogFormatSetter(json bool) {
	if !json {
		log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
			w.TimeFormat = time.RFC3339
		}))
	}
}
