package service

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"boilerplate/internal/model"
	"boilerplate/internal/repository"
)

// UserService handles business logic for users.
type UserService struct {
	repo   *repository.UserRepository
	jwtSec string
	jwtExp time.Duration
}

// NewUserService creates a new UserService.
func NewUserService(repo *repository.UserRepository, jwtSecret string, jwtExpiration time.Duration) *UserService {
	return &UserService{
		repo:   repo,
		jwtSec: jwtSecret,
		jwtExp: jwtExpiration,
	}
}

// Register creates a new user with hashed password.
func (s *UserService) Register(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hash),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

// Login authenticates a user and returns a JWT token.
func (s *UserService) Login(ctx context.Context, req *model.LoginRequest) (string, *model.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", nil, ErrAccountDisabled
	}

	token, err := generateJWT(user.ID, user.Email, s.jwtSec, s.jwtExp)
	if err != nil {
		return "", nil, err
	}

	resp := user.ToResponse()
	return token, &resp, nil
}

// GetByID fetches a user by ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*model.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

// List returns paginated users.
func (s *UserService) List(ctx context.Context, page, perPage int) ([]*model.UserResponse, int, error) {
	users, total, err := s.repo.List(ctx, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*model.UserResponse, len(users))
	for i, u := range users {
		r := u.ToResponse()
		resp[i] = &r
	}

	return resp, total, nil
}

// Update applies partial updates.
func (s *UserService) Update(ctx context.Context, id int64, req *model.UpdateUserRequest) (*model.UserResponse, error) {
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

// Delete removes a user.
func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
