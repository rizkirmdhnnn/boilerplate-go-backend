package handler

import (
	"errors"
	"net/http"
	"strconv"

	"boilerplate/internal/model"
	"boilerplate/internal/repository"
	"boilerplate/internal/service"
	"boilerplate/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register godoc
// @Summary Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.CreateUserRequest true "User registration"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Error(c, http.StatusConflict, "email already registered")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.Created(c, resp)
}

// Login godoc
// @Summary Authenticate a user
// @Tags users
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "Login credentials"
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	token, userResp, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrAccountDisabled) {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{
		"token": token,
		"user":  userResp,
	})
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		response.Unauthorized(c, "user not found in context")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// ListUsers godoc
// @Summary List users with pagination
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} response.APIResponse
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}

	users, total, err := h.svc.List(c.Request.Context(), page, perPage)
	if err != nil {
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.Paginated(c, users, page, perPage, total)
}

// GetUser godoc
// @Summary Get a user by ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	resp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// UpdateUser godoc
// @Summary Update a user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body model.UpdateUserRequest true "User update fields"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	resp, err := h.svc.Update(c.Request.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		if errors.Is(err, repository.ErrDuplicate) {
			response.Error(c, http.StatusConflict, "email already exists")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.OK(c, resp)
}

// DeleteUser godoc
// @Summary Delete a user
// @Tags users
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		_ = c.Error(err)
		response.InternalError(c)
		return
	}

	response.Message(c, "user deleted successfully")
}
