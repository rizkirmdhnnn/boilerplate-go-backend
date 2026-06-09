package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rizkirmdhnnn/boilerplate-go-backend/internal/domain"
	"github.com/rizkirmdhnnn/boilerplate-go-backend/pkg/database"
)

// UserRepository implements the domain.UserRepository port using PostgreSQL + pgx.
type UserRepository struct {
	db *database.Pool
}

// NewUserRepository creates a new PostgreSQL-backed UserRepository.
func NewUserRepository(db *database.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (email, name, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRow(ctx, query, user.Email, user.Name, user.Password, true)
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("%w: email already exists", domain.ErrDuplicate)
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `SELECT id, email, name, password, is_active, created_at, updated_at
	          FROM users WHERE id = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email (used for login).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, name, password, is_active, created_at, updated_at
	          FROM users WHERE email = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

// List retrieves users with pagination.
func (r *UserRepository) List(ctx context.Context, page, perPage int) ([]*domain.User, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if total == 0 {
		return []*domain.User{}, 0, nil
	}

	offset := (page - 1) * perPage
	query := `SELECT id, email, name, password, is_active, created_at, updated_at
	          FROM users ORDER BY id DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Password,
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// Update applies partial updates to a user.
func (r *UserRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	query := `UPDATE users SET `
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		query += fmt.Sprintf("%s = $%d, ", col, i)
		args = append(args, val)
		i++
	}
	query += fmt.Sprintf("updated_at = NOW() WHERE id = $%d", i)
	args = append(args, id)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("%w: email already exists", domain.ErrDuplicate)
		}
		return fmt.Errorf("update user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes a user by ID (hard delete).
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
