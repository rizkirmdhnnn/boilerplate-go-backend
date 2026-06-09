package repository

import "errors"

// Sentinel errors for repository operations.
var (
	ErrNotFound  = errors.New("resource not found")
	ErrDuplicate = errors.New("duplicate resource")
)
