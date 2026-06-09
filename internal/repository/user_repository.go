package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"boilerplate/internal/model"
	"boilerplate/pkg/database"
)

// UserRepository handles user data access.
type UserRepository struct {
	db *database.Pool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *database.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user. Returns the created user with ID.
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, name, password, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRow(ctx, query, user.Email, user.Name, user.Password, true)
	if err := row.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt); err != nil {
		// handle unique constraint violation
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("email already exists: %w", ErrDuplicate)
		}
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, email, name, password, is_active, created_at, updated_at 
	          FROM users WHERE id = $1`

	var user model.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email (used for login).
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, name, password, is_active, created_at, updated_at 
	          FROM users WHERE email = $1`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &user, nil
}

// List retrieves users with pagination.
func (r *UserRepository) List(ctx context.Context, page, perPage int) ([]*model.User, int, error) {
	// count total
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if total == 0 {
		return []*model.User{}, 0, nil
	}

	offset := (page - 1) * perPage
	query := `SELECT id, email, name, password, is_active, created_at, updated_at 
	          FROM users ORDER BY id DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var user model.User
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

	// add updated_at
	updates["updated_at"] = "NOW()"

	query := `UPDATE users SET `
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		if col == "updated_at" {
			query += fmt.Sprintf("%s = NOW(), ", col)
		} else {
			query += fmt.Sprintf("%s = $%d, ", col, i)
			args = append(args, val)
			i++
		}
	}
	query = query[:len(query)-2] // trim trailing ", "
	query += fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)

	tag, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("email already exists: %w", ErrDuplicate)
		}
		return fmt.Errorf("update user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
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
		return ErrNotFound
	}
	return nil
}
