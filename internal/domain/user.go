package domain

import (
	"context"
	"time"
)

// User represents the core domain entity.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  string    `json:"-"` // never exposed
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRepository is the port (interface) for data access.
// The domain defines this; infrastructure implements it.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, page, perPage int) ([]*User, int, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error
}
