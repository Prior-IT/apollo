package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"
)

/**
 * DOMAIN
 */

type User struct {
	ID     UserID
	Name   string
	Email  EmailAddress
	Joined time.Time
}

type (
	UserID uint
)

// NewUserID parses a user id from any unsigned integer.
func NewUserID(id uint) (UserID, error) {
	if id == 0 {
		return 0, errors.New("UserID cannot be 0")
	}
	return UserID(id), nil
}

// ParseUserID parses a string into a user id.
func ParseUserID(id string) (UserID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse user id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse user id: user ids cannot be negative")
	}
	userID, err := NewUserID(uint(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse user id: %w", err)
	}
	return userID, nil
}

/**
 * APPLICATION
 */

var ErrUserDoesNotExist = errors.New("that user does not exist")

type UserCreateData struct {
	Name  string
	Email EmailAddress
}

type UserService interface {
	// Create a new user with the specified data.
	CreateUser(ctx context.Context, data UserCreateData) (*User, error)
	// Retrieve the user with the specified id or ErrUserDoesNotExist if no such user exists.
	GetUser(ctx context.Context, id UserID) (*User, error)
	// Retrieve all existing users.
	ListUsers(ctx context.Context) ([]*User, error)
	// Retrieve the amount of existing users.
	GetAmountOfUsers(ctx context.Context) (uint64, error)
	// Delete the user with the specified id or ErrUserDoesNotExist if no such user exists.
	DeleteUser(ctx context.Context, id UserID) error
}
