package application

import (
	"context"
	"errors"
)

// Sentinel errors for the application / use-case layer.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is disabled")
)

// UserService defines the use-case interface for users.
// Handlers depend on this interface, not on the concrete implementation.
type UserService interface {
	Register(ctx context.Context, req *RegisterRequest) (*UserResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetByID(ctx context.Context, id int64) (*UserResponse, error)
	List(ctx context.Context, page, perPage int) ([]*UserResponse, int, error)
	Update(ctx context.Context, id int64, req *UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id int64) error
}
