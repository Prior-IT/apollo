package core

import "errors"

var (
	ErrUserDoesNotExist         = errors.New("user does not exist")
	ErrOrganisationDoesNotExist = errors.New("organisation does not exist")
	ErrNoActiveOrganisation     = errors.New("user has not chosen an organisation")
	ErrNotFound                 = errors.New("not found")
	ErrUnauthenticated          = errors.New("authentication required")
	ErrForbidden                = errors.New("user is not authorized")
	ErrConflict                 = errors.New("conflict")
)
