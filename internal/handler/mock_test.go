package handler

import (
	"context"
	"errors"
	"time"

	"github.com/stretchr/testify/mock"

	"boilerplate/internal/application"
)

// mockUserSvc implements application.UserService for testing.
type mockUserSvc struct {
	mock.Mock
}

func (m *mockUserSvc) Register(ctx context.Context, req *application.RegisterRequest) (*application.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.UserResponse), args.Error(1)
}

func (m *mockUserSvc) Login(ctx context.Context, req *application.LoginRequest) (*application.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.AuthResponse), args.Error(1)
}

func (m *mockUserSvc) GetByID(ctx context.Context, id int64) (*application.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.UserResponse), args.Error(1)
}

func (m *mockUserSvc) List(ctx context.Context, page, perPage int) ([]*application.UserResponse, int, error) {
	args := m.Called(ctx, page, perPage)
	users, _ := args.Get(0).([]*application.UserResponse)
	return users, args.Int(1), args.Error(2)
}

func (m *mockUserSvc) Update(ctx context.Context, id int64, req *application.UpdateUserRequest) (*application.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*application.UserResponse), args.Error(1)
}

func (m *mockUserSvc) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

var (
	errExample    = errors.New("something went wrong")
	errBadRequest = errors.New("bad request")
)

func dummyUserResp() *application.UserResponse {
	return &application.UserResponse{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		IsActive:  true,
		CreatedAt: time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC),
	}
}
