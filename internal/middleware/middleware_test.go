package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/application"
)

const testJWTSecret = "test-secret-key-for-middleware-testing"

// generateTestToken creates a valid JWT for testing.
func generateTestToken() string {
	claims := application.Claims{
		UserID: 1,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func TestCORS_AllowsAllOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_AllowsSpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"https://myapp.com"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://myapp.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, "https://myapp.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_DenyUnlistedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"https://myapp.com"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	r.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CORS([]string{"*"}))
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	r.ServeHTTP(w, req)

	assert.Equal(t, 204, w.Code)
}

func TestRequestID_ExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		c.String(200, id.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	r.ServeHTTP(w, req)

	assert.Equal(t, "my-custom-id", w.Body.String())
	assert.Equal(t, "my-custom-id", w.Header().Get("X-Request-ID"))
}

func TestRequestID_Generated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/test", func(c *gin.Context) {
		id, _ := c.Get("request_id")
		c.String(200, id.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.NotEmpty(t, w.Body.String())
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestAuthRequired_NoHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthRequired(testJWTSecret))
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthRequired(testJWTSecret))
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "NotBearer token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthRequired(testJWTSecret))
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-here")
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestAuthRequired_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	token := generateTestToken()

	r := gin.New()
	r.Use(AuthRequired(testJWTSecret))
	r.GET("/protected", func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		email := c.GetString("user_email")
		c.String(200, "user:%d,email:%s", userID, email)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user:1")
	assert.Contains(t, w.Body.String(), "test@example.com")
}

func TestAuthRequired_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Token signed with a different secret
	claims := application.Claims{
		UserID: 1,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongSecret, _ := token.SignedString([]byte("different-secret"))

	r := gin.New()
	r.Use(AuthRequired(testJWTSecret))
	r.GET("/protected", func(c *gin.Context) {
		c.String(200, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+wrongSecret)
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestLogLevelSetter(t *testing.T) {
	LogLevelSetter("debug")
	LogLevelSetter("info")
	LogLevelSetter("invalid") // should not panic
}

func TestLogFormatSetter(t *testing.T) {
	LogFormatSetter(true)  // JSON — no panic
	LogFormatSetter(false) // console — no panic
}

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}
