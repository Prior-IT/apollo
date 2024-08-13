package core

import "errors"

var (
	ErrUserDoesNotExist = errors.New("user does not exist")
	ErrUnauthenticated  = errors.New("authentication required")
	ErrForbidden        = errors.New("user is not authorized")
)
