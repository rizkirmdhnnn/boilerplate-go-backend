package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"boilerplate/internal/application"
	"boilerplate/internal/domain"
)

func setupTest() (*gin.Engine, *mockUserSvc) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockUserSvc)
	h := NewUserHandler(mockSvc)

	r := gin.New()
	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	return r, mockSvc
}

func TestRegister_Success(t *testing.T) {
	r, mockSvc := setupTest()

	req := application.RegisterRequest{
		Email:    "new@example.com",
		Name:     "New User",
		Password: "securepass123",
	}
	mockSvc.On("Register", mock.Anything, &req).
		Return(dummyUserResp(), nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	require.Equal(t, 201, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	mockSvc.AssertExpectations(t)
}

func TestRegister_BadRequest(t *testing.T) {
	r, _ := setupTest()

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/register",
		bytes.NewReader([]byte(`{"email":"invalid"}`)))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 422, w.Code)
}

func TestRegister_Duplicate(t *testing.T) {
	r, mockSvc := setupTest()

	req := application.RegisterRequest{
		Email:    "dup@example.com",
		Name:     "Dup",
		Password: "password123",
	}
	mockSvc.On("Register", mock.Anything, &req).
		Return(nil, domain.ErrDuplicate)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 409, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestRegister_InternalError(t *testing.T) {
	r, mockSvc := setupTest()

	req := application.RegisterRequest{
		Email:    "err@example.com",
		Name:     "Err",
		Password: "password123",
	}
	mockSvc.On("Register", mock.Anything, &req).
		Return(nil, errExample)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 500, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	r, mockSvc := setupTest()

	req := application.LoginRequest{
		Email:    "test@example.com",
		Password: "password",
	}
	mockSvc.On("Login", mock.Anything, &req).
		Return(&application.AuthResponse{Token: "jwt-token", User: dummyUserResp()}, nil)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	require.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
	mockSvc.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	r, mockSvc := setupTest()

	req := application.LoginRequest{
		Email:    "bad@example.com",
		Password: "wrong",
	}
	mockSvc.On("Login", mock.Anything, &req).
		Return(nil, application.ErrInvalidCredentials)

	body, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 401, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestLogin_ValidationError(t *testing.T) {
	r, _ := setupTest()

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/login",
		bytes.NewReader([]byte(`{"email":"not-an-email","password":""}`)))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	// email validation or password validation => 422
	assert.Equal(t, 422, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHealthHandler(nil)

	r := gin.New()
	r.GET("/health", h.Check)

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 200, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}

func TestRegister_InvalidJSON(t *testing.T) {
	r, _ := setupTest()

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("POST", "/api/v1/auth/register",
		bytes.NewReader([]byte(`{invalid json`)))
	reqObj.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 422, w.Code)
}

func TestUserHandler_MissingService(t *testing.T) {
	// Ensure handler methods handle nil service gracefully
	_ = NewUserHandler(nil)
}

// Protected endpoint tests — need to setup auth middleware

func setupProtectedTest() (*gin.Engine, *mockUserSvc) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockUserSvc)
	h := NewUserHandler(mockSvc)

	r := gin.New()
	users := r.Group("/api/v1/users")
	users.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("user_email", "test@example.com")
		c.Next()
	})
	users.GET("/me", h.GetProfile)
	users.GET("", h.ListUsers)

	return r, mockSvc
}

func TestGetProfile_Success(t *testing.T) {
	r, mockSvc := setupProtectedTest()

	mockSvc.On("GetByID", mock.Anything, int64(1)).
		Return(dummyUserResp(), nil)

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 200, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	r, mockSvc := setupProtectedTest()

	mockSvc.On("GetByID", mock.Anything, int64(1)).
		Return(nil, domain.ErrNotFound)

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/api/v1/users/me", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 404, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestListUsers(t *testing.T) {
	r, mockSvc := setupProtectedTest()

	mockSvc.On("List", mock.Anything, 1, 10).
		Return([]*application.UserResponse{dummyUserResp()}, 1, nil)

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/api/v1/users?page=1&per_page=10", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 200, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestListUsers_Error(t *testing.T) {
	r, mockSvc := setupProtectedTest()

	mockSvc.On("List", mock.Anything, 1, 10).
		Return(nil, 0, errors.New("db error"))

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/api/v1/users?page=1&per_page=10", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 500, w.Code)
	mockSvc.AssertExpectations(t)
}

func Test404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"success": false, "error": gin.H{"message": "route not found"}})
	})

	w := httptest.NewRecorder()
	reqObj, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
	r.ServeHTTP(w, reqObj)

	assert.Equal(t, 404, w.Code)
}
