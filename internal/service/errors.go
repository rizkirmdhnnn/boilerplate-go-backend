package service

import "errors"

// Sentinel errors for service layer.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountDisabled    = errors.New("account is disabled")
)
