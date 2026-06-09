package domain

import (
	"testing"
	"time"
)

func TestUserEntity(t *testing.T) {
	now := time.Now()

	u := &User{
		ID:        1,
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  "supersecret",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if u.ID != 1 {
		t.Errorf("ID = %d, want 1", u.ID)
	}
	if u.Email != "test@example.com" {
		t.Errorf("Email = %q, want %q", u.Email, "test@example.com")
	}
	if u.Name != "Test User" {
		t.Errorf("Name = %q, want %q", u.Name, "Test User")
	}
	if !u.IsActive {
		t.Error("IsActive should be true")
	}

	// Password field has json:"-" — ensure it still holds data
	if u.Password != "supersecret" {
		t.Errorf("Password = %q, want %q", u.Password, "supersecret")
	}
}

func TestUserZeroValues(t *testing.T) {
	u := &User{}

	if u.ID != 0 {
		t.Error("zero-value User should have ID=0")
	}
	if u.Email != "" {
		t.Error("zero-value User should have empty email")
	}
	if u.IsActive {
		t.Error("zero-value User should have IsActive=false")
	}
}
