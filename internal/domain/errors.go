package domain

import "errors"

// Sentinel errors — domain layer defines them.
var (
	ErrNotFound  = errors.New("resource not found")
	ErrDuplicate = errors.New("duplicate resource")
)
