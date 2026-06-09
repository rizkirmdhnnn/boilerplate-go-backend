package domain

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrNotFound", ErrNotFound, "resource not found"},
		{"ErrDuplicate", ErrDuplicate, "duplicate resource"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	if errors.Is(ErrNotFound, ErrDuplicate) {
		t.Error("ErrNotFound should not be ErrDuplicate")
	}
	if errors.Is(ErrDuplicate, ErrNotFound) {
		t.Error("ErrDuplicate should not be ErrNotFound")
	}
}
