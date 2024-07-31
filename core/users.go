package core

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

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
func NewUserID(id uint32) (UserID, error) {
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
	userID, err := NewUserID(uint32(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse user id: %w", err)
	}
	return userID, nil
}
