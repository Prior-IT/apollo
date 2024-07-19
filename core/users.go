package core

import (
	"errors"
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
