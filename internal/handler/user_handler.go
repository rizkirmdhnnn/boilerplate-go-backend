package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"boilerplate/internal/application"
	"boilerplate/internal/domain"
	"boilerplate/pkg/cache"
	"boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests.
// Depends ONLY on the application.UserService interface (use-case port).
type UserHandler struct {
	svc   application.UserService
	cache cache.Cache
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc application.UserService, c cache.Cache) *UserHandler {
	return &UserHandler{svc: svc, cache: c}
}

// Register     godoc
// @Summary      Register a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body application.RegisterRequest true "User registration"
// @Success      201  {object}  response.APIResponse
// @Failure      400  {object}  response.APIResponse
// @Failure      409  {object}  response.APIResponse
// @Router       /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req application.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicate) {
			response.Error(c, http.StatusConflict, "email already registered")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.Created(c, resp)
}

// Login        godoc
// @Summary      Authenticate a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body application.LoginRequest true "Login credentials"
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Router       /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req application.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, application.ErrInvalidCredentials) ||
			errors.Is(err, application.ErrAccountDisabled) {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// GetProfile   godoc
// @Summary      Get current user profile
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Router       /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "user not found in context")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// ListUsers    godoc
// @Summary      List users with pagination (cached for 30s)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page     query  int  false  "Page number"  default(1)
// @Param        per_page query  int  false  "Items per page"  default(10)
// @Success      200  {object}  response.APIResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	// Try cache first
	cacheKey := fmt.Sprintf("users:list:p%d:pp%d", page, perPage)
	if cached, err := h.cache.Get(c.Request.Context(), cacheKey); err == nil {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(cached))
		return
	}

	users, total, err := h.svc.List(c.Request.Context(), page, perPage)
	if err != nil {
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	resp := response.APIResponse{
		Success: true,
		Data:    users,
		Meta: &response.Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	// Cache for 30s (best-effort, don't fail on cache error)
	if data, err := json.Marshal(resp); err == nil {
		_ = h.cache.Set(c.Request.Context(), cacheKey, string(data), 30*time.Second)
	}

	c.JSON(http.StatusOK, resp)
}

// GetUser      godoc
// @Summary      Get a user by ID
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// UpdateUser   godoc
// @Summary      Update a user
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Param        body body      application.UpdateUserRequest true "User update fields"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req application.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		if errors.Is(err, domain.ErrDuplicate) {
			response.Error(c, http.StatusConflict, "email already exists")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// DeleteUser   godoc
// @Summary      Delete a user
// @Tags         users
// @Security     BearerAuth
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.Message(c, "user deleted successfully")
}
