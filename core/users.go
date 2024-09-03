package core

import (
	"context"
	"time"
)

/**
 * DOMAIN
 */

type User struct {
	ID     UserID
	Name   string
	Email  EmailAddress
	Admin  bool
	Joined time.Time
}

type UserID ID

/**
 * APPLICATION
 */

type UserUpdate struct {
	Name  *string
	Email *string
}

type UserService interface {
	// Create a new user with the specified data.
	CreateUser(ctx context.Context, name string, email EmailAddress) (*User, error)
	// Retrieve the user with the specified id or ErrUserDoesNotExist if no such user exists.
	GetUser(ctx context.Context, id UserID) (*User, error)
	// Retrieve all existing users.
	ListUsers(ctx context.Context) ([]User, error)
	// Retrieve the amount of existing users.
	GetAmountOfUsers(ctx context.Context) (uint64, error)
	// Delete the user with the specified id or ErrUserDoesNotExist if no such user exists.
	DeleteUser(ctx context.Context, id UserID) error
	// Update the user's admin state to the specified state.
	UpdateUserAdmin(ctx context.Context, id UserID, admin bool) error
	// Update the user with the specified data.
	UpdateUser(ctx context.Context, id UserID, data UserUpdate) (*User, error)
}
