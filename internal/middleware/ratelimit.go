package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a sliding-window in-memory rate limiter per IP.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	rate     int           // max requests per minute
	burst    int           // burst size (additional allowance)
	ticker   *time.Ticker
	stopCh   chan struct{}
}

type visitor struct {
	count     int
	resetTime time.Time
}

// RateLimiterOption configures the rate limiter.
type RateLimiterOption func(*RateLimiter)

// WithRate sets the max requests per minute (default: 100).
func WithRate(rate int) RateLimiterOption {
	return func(rl *RateLimiter) {
		if rate > 0 {
			rl.rate = rate
		}
	}
}

// WithBurst sets the burst size (default: 20).
func WithBurst(burst int) RateLimiterOption {
	return func(rl *RateLimiter) {
		if burst >= 0 {
			rl.burst = burst
		}
	}
}

// WithCleanupInterval sets how often stale entries are purged (default: 1 minute).
func WithCleanupInterval(d time.Duration) RateLimiterOption {
	return func(rl *RateLimiter) {
		if d > 0 {
			rl.ticker.Stop()
			rl.ticker = time.NewTicker(d)
		}
	}
}

// NewRateLimiter creates a new RateLimiter with the given options.
// Defaults: 100 requests/min, 20 burst.
func NewRateLimiter(opts ...RateLimiterOption) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     100,
		burst:    20,
		ticker:   time.NewTicker(1 * time.Minute),
		stopCh:   make(chan struct{}),
	}

	for _, opt := range opts {
		opt(rl)
	}

	go rl.cleanupLoop()

	return rl
}

// NewRateLimiterWithParams is a convenience constructor for simple usage.
func NewRateLimiterWithParams(rate, burst int) *RateLimiter {
	return NewRateLimiter(WithRate(rate), WithBurst(burst))
}

// cleanupLoop periodically purges stale visitor entries.
func (rl *RateLimiter) cleanupLoop() {
	for {
		select {
		case <-rl.ticker.C:
			rl.cleanup()
		case <-rl.stopCh:
			rl.ticker.Stop()
			return
		}
	}
}

// cleanup removes visitors whose window has expired.
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, v := range rl.visitors {
		if now.After(v.resetTime) {
			delete(rl.visitors, ip)
		}
	}
}

// Stop shuts down the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// allow checks whether a request from the given IP should be allowed.
func (rl *RateLimiter) allow(ip string) (bool, int, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	v, exists := rl.visitors[ip]
	if !exists || now.After(v.resetTime) {
		// New window, allow up to rate + burst
		resetTime := now.Add(1 * time.Minute).Truncate(time.Second)
		rl.visitors[ip] = &visitor{
			count:     1,
			resetTime: resetTime,
		}
		return true, rl.rate + rl.burst - 1, rl.rate + rl.burst, resetTime
	}

	allowed := rl.rate + rl.burst
	remaining := allowed - v.count
	if remaining < 0 {
		remaining = 0
	}

	if v.count >= allowed {
		return false, 0, allowed, v.resetTime
	}

	v.count++
	return true, remaining - 1, allowed, v.resetTime
}

// RateLimit returns a Gin handler that enforces rate limiting per IP.
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		ok, remaining, limit, resetTime := rl.allow(ip)

		// Always set headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.Itoa(int(resetTime.Unix())))

		if !ok {
			response.Error(c, http.StatusTooManyRequests, "too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
