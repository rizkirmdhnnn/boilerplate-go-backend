package model

import "time"

// User represents a user in the system.
type User struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email" binding:"required,email"`
	Name      string    `json:"name" db:"name" binding:"required,min=2,max=100"`
	Password  string    `json:"-" db:"password"` // never expose in JSON
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserResponse is the public-facing user payload (no password).
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a User to a safe response.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// CreateUserRequest is the input for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest is the input for updating a user.
type UpdateUserRequest struct {
	Email *string `json:"email" binding:"omitempty,email"`
	Name  *string `json:"name" binding:"omitempty,min=2,max=100"`
}

// LoginRequest is the input for authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
