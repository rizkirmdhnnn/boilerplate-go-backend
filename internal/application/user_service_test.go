package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"boilerplate/internal/domain"
)

const testJWTSecret = "test-secret-key-that-is-long-enough-for-testing"

func TestRegister_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "new@example.com" && u.Name == "New User"
	})).Return(nil).Run(func(args mock.Arguments) {
		u := args.Get(1).(*domain.User)
		u.ID = 1
		u.CreatedAt = fixedTime()
		u.UpdatedAt = fixedTime()
	})

	resp, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "new@example.com",
		Name:     "New User",
		Password: "securepass123",
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "new@example.com", resp.Email)
	assert.Equal(t, "New User", resp.Name)
	assert.True(t, resp.IsActive)
	mockRepo.AssertExpectations(t)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("Create", mock.Anything, mock.Anything).
		Return(domain.ErrDuplicate)

	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "dup@example.com",
		Name:     "Dup",
		Password: "securepass123",
	})

	assert.ErrorIs(t, err, domain.ErrDuplicate)
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	// Create a user with a real bcrypt password hash
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	user := dummyUser()
	user.Email = "login@example.com"
	user.Password = string(hash)

	mockRepo.On("GetByEmail", mock.Anything, "login@example.com").
		Return(user, nil)

	authResp, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "login@example.com",
		Password: "correctpassword",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, authResp.Token)
	assert.Equal(t, user.ID, authResp.User.ID)
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.MinCost)
	user := dummyUser()
	user.Email = "login@example.com"
	user.Password = string(hash)

	mockRepo.On("GetByEmail", mock.Anything, "login@example.com").
		Return(user, nil)

	_, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "login@example.com",
		Password: "wrongpassword",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").
		Return(nil, domain.ErrNotFound)

	_, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "anypassword",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	mockRepo.AssertExpectations(t)
}

func TestLogin_AccountDisabled(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	user := dummyUser()
	user.Email = "disabled@example.com"
	user.Password = string(hash)
	user.IsActive = false

	mockRepo.On("GetByEmail", mock.Anything, "disabled@example.com").
		Return(user, nil)

	_, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "disabled@example.com",
		Password: "password",
	})

	assert.ErrorIs(t, err, ErrAccountDisabled)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	user := dummyUser()
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(user, nil)

	resp, err := svc.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.Name, resp.Name)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("GetByID", mock.Anything, int64(999)).
		Return(nil, domain.ErrNotFound)

	_, err := svc.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestList_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	users := []*domain.User{dummyUser(), dummyUser()}
	mockRepo.On("List", mock.Anything, 1, 10).Return(users, 2, nil)

	resp, total, err := svc.List(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, 2, total)
	mockRepo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("List", mock.Anything, 1, 10).
		Return([]*domain.User{}, 0, nil)

	resp, total, err := svc.List(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, resp)
	assert.Equal(t, 0, total)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	newName := "Updated Name"
	mockRepo.On("Update", mock.Anything, int64(1),
		mock.MatchedBy(func(m map[string]interface{}) bool {
			return m["name"] == "Updated Name"
		})).Return(nil)

	// After update, GetByID gets called to return the fresh user
	updated := dummyUser()
	updated.Name = newName
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(updated, nil)

	resp, err := svc.Update(context.Background(), 1, &UpdateUserRequest{
		Name: &newName,
	})

	require.NoError(t, err)
	assert.Equal(t, newName, resp.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	name := "Any"
	mockRepo.On("Update", mock.Anything, int64(999), mock.Anything).
		Return(domain.ErrNotFound)

	_, err := svc.Update(context.Background(), 999, &UpdateUserRequest{
		Name: &name,
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	mockRepo.On("Delete", mock.Anything, int64(999)).
		Return(domain.ErrNotFound)

	err := svc.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, domain.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestLogin_GeneratesValidJWT(t *testing.T) {
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.MinCost)
	user := dummyUser()
	user.Password = string(hash)

	mockRepo.On("GetByEmail", mock.Anything, "test@example.com").
		Return(user, nil)

	authResp, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "test@example.com",
		Password: "pass",
	})
	require.NoError(t, err)

	// Verify the token can be validated
	claims, err := ValidateJWT(authResp.Token, testJWTSecret)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	mockRepo.AssertExpectations(t)
}

func TestRegister_ErrorsNotWrapped(t *testing.T) {
	// Ensure random errors are passed through, not absorbed
	mockRepo := new(mockUserRepo)
	svc := NewUserService(mockRepo, testJWTSecret, time.Hour)

	unexpected := errors.New("connection reset by peer")
	mockRepo.On("Create", mock.Anything, mock.Anything).
		Return(unexpected)

	_, err := svc.Register(context.Background(), &RegisterRequest{
		Email:    "x@y.com",
		Name:     "X",
		Password: "password123",
	})
	assert.ErrorIs(t, err, unexpected)
	mockRepo.AssertExpectations(t)
}
