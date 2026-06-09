package application

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/domain"
)

// mockUserRepo implements domain.UserRepository for testing.
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) List(ctx context.Context, page, perPage int) ([]*domain.User, int, error) {
	args := m.Called(ctx, page, perPage)
	users, _ := args.Get(0).([]*domain.User)
	return users, args.Int(1), args.Error(2)
}

func (m *mockUserRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// test helpers

func fixedTime() time.Time {
	return time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)
}

func dummyUser() *domain.User {
	t := fixedTime()
	return &domain.User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "$2a$10$hashedpassword",
		IsActive:  true,
		CreatedAt: t,
		UpdatedAt: t,
	}
}
