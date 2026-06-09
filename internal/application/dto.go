package application

import "time"

// --- Request DTOs ---

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// LoginRequest is the input for authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UpdateUserRequest is the input for updating a user.
type UpdateUserRequest struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
}

// --- Response DTOs ---

// UserResponse is the public-facing user payload (no password).
type UserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthResponse wraps a JWT token and user data.
type AuthResponse struct {
	Token string       `json:"token"`
	User  *UserResponse `json:"user"`
}
