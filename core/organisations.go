package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

/**
 * DOMAIN
 */

type Organisation struct {
	ID   OrganisationID
	Name string
}

type (
	OrganisationID uint
)

func (id OrganisationID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// NewOrganisationID parses an organisation id from any unsigned integer.
func NewOrganisationID(id uint) (OrganisationID, error) {
	if id == 0 {
		return 0, errors.New("OrganisationID cannot be 0")
	}
	return OrganisationID(id), nil
}

// ParseOrganisationID parses a string into an organisation id.
func ParseOrganisationID(id string) (OrganisationID, error) {
	integerID, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("cannot parse organisation id: %w", err)
	}
	if integerID < 0 {
		return 0, errors.New("cannot parse organisation id: organisation ids cannot be negative")
	}
	organisationID, err := NewOrganisationID(uint(integerID))
	if err != nil {
		return 0, fmt.Errorf("cannot parse organisation id: %w", err)
	}
	return organisationID, nil
}

/**
 * APPLICATION
 */

type OrganisationCreateData struct {
	Name string
}

type OrganisationService interface {
	// Create a new organisation with the specified data.
	CreateOrganisation(ctx context.Context, data OrganisationCreateData) (*Organisation, error)
	// Retrieve the organisation with the specified id or ErrOrganisationDoesNotExist if no such organisation exists.
	GetOrganisation(ctx context.Context, id OrganisationID) (*Organisation, error)
	// Retrieve all existing organisations.
	ListOrganisations(ctx context.Context) ([]Organisation, error)
	// Retrieve the amount of existing organisations.
	GetAmountOfOrganisations(ctx context.Context) (uint64, error)
	// Delete the organisation with the specified id or ErrOrganisationDoesNotExist if no such organisation exists.
	DeleteOrganisation(ctx context.Context, id OrganisationID) error
	// List the organisations a user belongs to or ErrUserDoesNotExist if no such user exists
	ListOrganisationsForUser(ctx context.Context, id UserID) ([]Organisation, error)
	// List the users that belong to an organisation or ErrOrganisationDoesNotExist if no such organisation exists
	ListUsersinOrganisation(ctx context.Context, id OrganisationID) ([]User, error)
	// Add user to an existing organisation
	AddUserToOrganisation(ctx context.Context, UserID UserID, OrgID OrganisationID) (*Organisation, error)
	// Remove user from an organisation
	RemoveUserFromOrganisation(ctx context.Context, UserID UserID, OrgID OrganisationID) (*Organisation, error)
}
