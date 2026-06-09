package application

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/domain"
)

// userService implements the UserService use-case interface.
type userService struct {
	repo   domain.UserRepository
	jwtSec string
	jwtExp time.Duration
}

// NewUserService creates a concrete UserService.
// Accepts a domain.UserRepository (port) — any adapter can be injected.
func NewUserService(repo domain.UserRepository, jwtSecret string, jwtExp time.Duration) UserService {
	return &userService{
		repo:   repo,
		jwtSec: jwtSecret,
		jwtExp: jwtExp,
	}
}

func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hash),
		IsActive: true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *userService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	token, err := generateJWT(user.ID, user.Email, s.jwtSec, s.jwtExp)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *userService) GetByID(ctx context.Context, id int64) (*UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *userService) List(ctx context.Context, page, perPage int) ([]*UserResponse, int, error) {
	users, total, err := s.repo.List(ctx, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*UserResponse, len(users))
	for i, u := range users {
		resp[i] = toUserResponse(u)
	}

	return resp, total, nil
}

func (s *userService) Update(ctx context.Context, id int64, req *UpdateUserRequest) (*UserResponse, error) {
	updates := make(map[string]interface{})
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *userService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// toUserResponse converts a domain User to a safe response DTO.
func toUserResponse(u *domain.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
